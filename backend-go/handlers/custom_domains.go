package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nrednav/cuid2"

	"boop-cat/db"
	"boop-cat/deploy"
	"boop-cat/middleware"
)

type CustomDomainHandler struct {
	DB     *sql.DB
	Engine *deploy.Engine
}

func NewCustomDomainHandler(database *sql.DB, engine *deploy.Engine) *CustomDomainHandler {
	return &CustomDomainHandler{DB: database, Engine: engine}
}

func (h *CustomDomainHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireLogin)

	r.Get("/sites/{siteId}/custom-domains", h.ListCustomDomains)
	r.Post("/sites/{siteId}/custom-domains", h.CreateCustomDomain)
	r.Delete("/sites/{siteId}/custom-domains/{id}", h.DeleteCustomDomain)

	r.Post("/sites/{siteId}/custom-domains/{id}/poll", h.PollCustomDomain)

	return r
}

func (h *CustomDomainHandler) ListCustomDomains(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "siteId")

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	domains, err := db.ListCustomDomains(h.DB, siteID)
	if err != nil {
		jsonError(w, "list-failed", http.StatusInternalServerError)
		return
	}

	resp := []db.CustomDomainResponse{}
	for _, d := range domains {
		resp = append(resp, d.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CustomDomainHandler) CreateCustomDomain(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "siteId")

	var req struct {
		Hostname string `json:"hostname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid-json", http.StatusBadRequest)
		return
	}

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	if req.Hostname == "" {
		jsonError(w, "hostname-required", http.StatusBadRequest)
		return
	}
	hostname := strings.ToLower(strings.TrimSpace(req.Hostname))
	rootDomain := strings.ToLower(os.Getenv("FSD_EDGE_ROOT_DOMAIN"))
	if rootDomain != "" && (hostname == rootDomain || strings.HasSuffix(hostname, "."+rootDomain)) {
		jsonError(w, "root-domains-not-supported", http.StatusBadRequest)
		return
	}

	count, err := db.CountCustomDomainsForUser(h.DB, userID)
	if err != nil {
		jsonError(w, "db-error", http.StatusInternalServerError)
		return
	}
	if count >= 3 {
		jsonError(w, "custom-domain-limit-reached", http.StatusBadRequest)
		return
	}

	cf := deploy.NewCloudflareClient(h.Engine.CFAccountID, h.Engine.CFNamespaceID, h.Engine.CFToken)

	zoneID, err := cf.GetZoneID(os.Getenv("FSD_EDGE_ROOT_DOMAIN"))
	if err != nil {

		jsonError(w, "zone-not-found", http.StatusInternalServerError)
		return
	}

	if err := h.ensureFallbackOrigin(cf, zoneID); err != nil {
		fmt.Printf("Fallback origin error: %v\n", err)

	}

	res, err := cf.CreateCustomHostname(zoneID, hostname)
	if err != nil {
		jsonError(w, "custom-domain-create-failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	var cfRes CloudflareHostname
	if err := json.Unmarshal(res, &cfRes); err != nil {
		fmt.Printf("Error unmarshalling CF response: %v\n", err)
	}

	sslStatus := cfRes.SSL.Status
	status := cfRes.Status
	combined := "pending"
	if status == "active" && sslStatus == "active" {
		combined = "active"
	} else if status == "active" || sslStatus == "active" {
		combined = "pending_ssl"
	}

	records := extractVerificationRecords(cfRes)
	recordsJSON, _ := json.Marshal(records)

	id := cuid2.Generate()
	err = db.CreateCustomDomain(h.DB, id, siteID, hostname, cfRes.ID, combined, sslStatus, string(recordsJSON))
	if err != nil {
		jsonError(w, "db-create-failed", http.StatusInternalServerError)
		return
	}

	if combined == "active" {
		cf.EnsureRouting(hostname, siteID, "")
	}

	d, _ := db.GetCustomDomainByID(h.DB, id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d.ToResponse())
}

func (h *CustomDomainHandler) PollCustomDomain(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "siteId")
	id := chi.URLParam(r, "id")

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	domain, err := db.GetCustomDomainByID(h.DB, id)
	if err != nil {
		jsonError(w, "custom-domain-not-found", http.StatusNotFound)
		return
	}

	cf := deploy.NewCloudflareClient(h.Engine.CFAccountID, h.Engine.CFNamespaceID, h.Engine.CFToken)
	zoneID, _ := cf.GetZoneID(os.Getenv("FSD_EDGE_ROOT_DOMAIN"))

	res, err := cf.GetCustomHostname(zoneID, domain.CFCustomHostnameID.String)
	if err != nil {
		jsonError(w, "poll-failed", http.StatusInternalServerError)
		return
	}

	var cfRes CloudflareHostname
	if err := json.Unmarshal(res, &cfRes); err != nil {
		fmt.Printf("Error unmarshalling CF response: %v\n", err)
	}

	sslStatus := cfRes.SSL.Status
	status := cfRes.Status
	combined := "pending"
	if status == "active" && sslStatus == "active" {
		combined = "active"
	} else if status == "active" || sslStatus == "active" {
		combined = "pending_ssl"
	}

	records := extractVerificationRecords(cfRes)
	recordsJSON, _ := json.Marshal(records)

	db.UpdateCustomDomainStatus(h.DB, id, combined, sslStatus, string(recordsJSON), cfRes.ID)

	if combined == "active" && domain.Status != "active" {
		cf.EnsureRouting(domain.Hostname, siteID, "")
	}

	updated, _ := db.GetCustomDomainByID(h.DB, id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated.ToResponse())
}

type CloudflareHostname struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	SSL    struct {
		Status            string `json:"status"`
		ValidationRecords []struct {
			TxtName  string `json:"txt_name"`
			TxtValue string `json:"txt_value"`
			HTTPUrl  string `json:"http_url"`
			HTTPBody string `json:"http_body"`
		} `json:"validation_records"`
	} `json:"ssl"`
	OwnershipVerification struct {
		Type  string `json:"type"`
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"ownership_verification"`
	OwnershipVerificationHTTP struct {
		HTTPUrl  string `json:"http_url"`
		HTTPBody string `json:"http_body"`
	} `json:"ownership_verification_http"`
}

type VerificationRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	HTTPUrl  string `json:"http_url,omitempty"`
	HTTPBody string `json:"http_body,omitempty"`
}

func extractVerificationRecords(r CloudflareHostname) []VerificationRecord {
	var records []VerificationRecord

	if r.OwnershipVerification.Type != "" && r.OwnershipVerification.Name != "" && r.OwnershipVerification.Value != "" {
		records = append(records, VerificationRecord{
			Type:  r.OwnershipVerification.Type,
			Name:  r.OwnershipVerification.Name,
			Value: r.OwnershipVerification.Value,
		})
	}

	if r.OwnershipVerificationHTTP.HTTPUrl != "" && r.OwnershipVerificationHTTP.HTTPBody != "" {
		records = append(records, VerificationRecord{
			Type:     "http",
			HTTPUrl:  r.OwnershipVerificationHTTP.HTTPUrl,
			HTTPBody: r.OwnershipVerificationHTTP.HTTPBody,
		})
	}

	for _, rec := range r.SSL.ValidationRecords {
		if rec.TxtName != "" && rec.TxtValue != "" {
			records = append(records, VerificationRecord{
				Type:  "ssl_txt",
				Name:  rec.TxtName,
				Value: rec.TxtValue,
			})
		}
		if rec.HTTPUrl != "" && rec.HTTPBody != "" {
			records = append(records, VerificationRecord{
				Type:     "ssl_http",
				HTTPUrl:  rec.HTTPUrl,
				HTTPBody: rec.HTTPBody,
			})
		}
	}

	return records
}

func (h *CustomDomainHandler) DeleteCustomDomain(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	siteID := chi.URLParam(r, "siteId")
	id := chi.URLParam(r, "id")

	site, err := db.GetSiteByID(h.DB, userID, siteID)
	if err != nil || site == nil {
		jsonError(w, "site-not-found", http.StatusNotFound)
		return
	}

	domain, err := db.GetCustomDomainByID(h.DB, id)
	if err != nil {
		jsonError(w, "custom-domain-not-found", http.StatusNotFound)
		return
	}

	cf := deploy.NewCloudflareClient(h.Engine.CFAccountID, h.Engine.CFNamespaceID, h.Engine.CFToken)
	zoneID, _ := cf.GetZoneID(os.Getenv("FSD_EDGE_ROOT_DOMAIN"))

	cfID := domain.CFCustomHostnameID.String
	if cfID == "" {

		foundID, err := cf.GetCustomHostnameIDByName(zoneID, domain.Hostname)
		if err == nil {
			cfID = foundID
		} else {
			fmt.Printf("Warning: Could not find CF ID for %s: %v\n", domain.Hostname, err)
		}
	}

	if cfID != "" {
		err := cf.DeleteCustomHostname(zoneID, cfID)
		if err != nil {
			fmt.Printf("Warning: Failed to delete custom hostname %s: %v\n", cfID, err)
		}
	}

	cf.RemoveRouting("", "", domain.Hostname)

	db.DeleteCustomDomain(h.DB, id)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (h *CustomDomainHandler) ensureFallbackOrigin(cf *deploy.CloudflareClient, zoneID string) error {
	rootDomain := os.Getenv("FSD_EDGE_ROOT_DOMAIN")
	if rootDomain == "" {
		return nil
	}

	fallbackOrigin := fmt.Sprintf("sites.%s", rootDomain)

	records, err := cf.GetDNSRecords(zoneID, fallbackOrigin)
	if err != nil {
		fmt.Printf("Error getting DNS records: %v\n", err)

	} else {
		var existing *deploy.DNSRecord
		for _, r := range records {
			if r.Name == fallbackOrigin {
				existing = &r
				break
			}
		}

		targetRecord := deploy.DNSRecord{
			Type: "AAAA",
			Name: "sites",

			Content: "100::",
			Proxied: true,
		}

		targetRecord.Name = fallbackOrigin

		if existing == nil {
			err := cf.CreateDNSRecord(zoneID, targetRecord)
			if err != nil {
				fmt.Printf("Error creating DNS record: %v\n", err)
			}
		} else {
			if existing.Type != "AAAA" || existing.Content != "100::" || !existing.Proxied {
				err := cf.UpdateDNSRecord(zoneID, existing.ID, targetRecord)
				if err != nil {
					fmt.Printf("Error updating DNS record: %v\n", err)
				}
			}
		}
	}

	fmt.Printf("Setting fallback origin for zone %s to %s\n", zoneID, fallbackOrigin)
	return cf.UpdateFallbackOrigin(zoneID, fallbackOrigin)
}
