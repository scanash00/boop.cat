package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type DMCAMonitor struct {
	IMAPHost   string
	IMAPPort   string
	IMAPUser   string
	IMAPPass   string
	IMAPTLS    bool
	WebhookURL string
	pollTicker *time.Ticker
	stopChan   chan struct{}
}

func NewDMCAMonitor() *DMCAMonitor {
	return &DMCAMonitor{
		IMAPHost:   os.Getenv("IMAP_HOST"),
		IMAPPort:   os.Getenv("IMAP_PORT"),
		IMAPUser:   os.Getenv("IMAP_USER"),
		IMAPPass:   os.Getenv("IMAP_PASSWORD"),
		IMAPTLS:    os.Getenv("IMAP_TLS") == "true",
		WebhookURL: os.Getenv("DISCORD_DMCA_WEBHOOK_URL"),
		stopChan:   make(chan struct{}),
	}
}

func (m *DMCAMonitor) Start() {
	if m.IMAPHost == "" || m.IMAPUser == "" || m.IMAPPass == "" {
		log.Println("DMCA Monitor: Missing IMAP credentials, skipping.")
		return
	}

	m.CheckEmails()

	m.pollTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-m.pollTicker.C:
				m.CheckEmails()
			case <-m.stopChan:
				m.pollTicker.Stop()
				return
			}
		}
	}()
	log.Println("DMCA Monitor started.")
}

func (m *DMCAMonitor) Stop() {
	close(m.stopChan)
}

func (m *DMCAMonitor) CheckEmails() {
	if m.IMAPHost == "" || m.IMAPUser == "" || m.IMAPPass == "" {
		log.Println("DMCA Monitor: Missing IMAP credentials, skipping.")
		return
	}

	port := m.IMAPPort
	if port == "" {
		port = "993"
	}

	addr := fmt.Sprintf("%s:%s", m.IMAPHost, port)

	var c *client.Client
	var err error

	if m.IMAPTLS {
		c, err = client.DialTLS(addr, nil)
	} else {
		c, err = client.Dial(addr)
		if err == nil {

			if ok, _ := c.SupportStartTLS(); ok {
				if err := c.StartTLS(nil); err != nil {
					log.Printf("DMCA Monitor: STARTTLS failed: %v", err)
					return
				}
			}
		}
	}
	if err != nil {
		log.Printf("DMCA Monitor: Failed to connect: %v", err)
		return
	}
	defer c.Logout()

	if err := c.Login(m.IMAPUser, m.IMAPPass); err != nil {
		log.Printf("DMCA Monitor: Login failed: %v", err)
		return
	}

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Printf("DMCA Monitor: Failed to select INBOX: %v", err)
		return
	}

	if mbox.Messages == 0 {
		return
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	uids, err := c.Search(criteria)
	if err != nil {
		log.Printf("DMCA Monitor: Search failed: %v", err)
		return
	}

	if len(uids) == 0 {
		return
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

	messages := make(chan *imap.Message, len(uids))
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	for msg := range messages {
		if msg == nil {
			continue
		}

		r := msg.GetBody(section)
		if r == nil {
			continue
		}

		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Printf("DMCA Monitor: Failed to create mail reader: %v", err)
			continue
		}

		var from, subject, body string

		header := mr.Header
		if addrs, err := header.AddressList("From"); err == nil && len(addrs) > 0 {
			from = addrs[0].String()
		}
		if s, err := header.Subject(); err == nil {
			subject = s
		}

		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:

				contentType, _, _ := h.ContentType()

				buf := new(bytes.Buffer)
				buf.ReadFrom(p.Body)
				partBody := buf.String()

				if body == "" {
					body = partBody
				} else if contentType == "text/plain" {

					body += "\n\n" + partBody
				}
			}
		}

		log.Printf("DMCA Monitor: Forwarding email from %s: %s", from, subject)

		if m.WebhookURL != "" {
			m.sendToDiscord(from, subject, body)
		} else {
			log.Println("DMCA Monitor: No Webhook URL configured.")
		}
	}

	if err := <-done; err != nil {
		log.Printf("DMCA Monitor: Fetch error: %v", err)
	}

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	if err := c.Store(seqSet, item, flags, nil); err != nil {
		log.Printf("DMCA Monitor: Store failed: %v", err)
	}
}

func (m *DMCAMonitor) sendToDiscord(from, subject, body string) {
	if m.WebhookURL == "" {
		return
	}

	header := fmt.Sprintf("**New DMCA/Legal-related Email Received**\n**From:** %s\n**Subject:** %s\n\n", from, subject)

	if body == "" {
		body = "(No content)"
	}

	const chunkSize = 1900
	chunks := splitString(body, chunkSize)

	firstChunk := header + ">>> " + chunks[0]
	m.postWebhook(firstChunk)

	for i := 1; i < len(chunks); i++ {
		time.Sleep(500 * time.Millisecond)
		m.postWebhook(">>> " + chunks[i])
	}
}

func (m *DMCAMonitor) postWebhook(content string) {
	payload := map[string]string{"content": content}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(m.WebhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("DMCA Monitor: Discord webhook failed: %v", err)
		return
	}
	resp.Body.Close()
}

func splitString(s string, chunkSize int) []string {
	if len(s) <= chunkSize {
		return []string{s}
	}

	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func StartDMCAMonitor() *DMCAMonitor {
	monitor := NewDMCAMonitor()
	monitor.Start()
	return monitor
}
