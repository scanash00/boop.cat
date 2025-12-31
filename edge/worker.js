// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

const ASSET_EXTENSIONS =
  /\.(js|mjs|css|png|jpg|jpeg|webp|avif|svg|gif|ico|woff|woff2|ttf|otf|eot|map|json|xml|txt|pdf|mp4|webm|mp3|wav)$/i;

function parseSubdomain(hostname, rootDomain) {
  if (!rootDomain) return null;
  const h = hostname.toLowerCase();
  const root = rootDomain.toLowerCase();
  if (h === root || !h.endsWith(`.${root}`)) return null;
  const sub = h.slice(0, -(root.length + 1));
  return !sub || sub.includes('.') ? null : sub;
}

function isAssetPath(pathname) {
  return pathname.startsWith('/assets/') || ASSET_EXTENSIONS.test(pathname);
}

function getCacheControl(pathname) {
  if (pathname === '/' || pathname.endsWith('.html')) {
    return 'public, max-age=60, s-maxage=60';
  }

  if (pathname.startsWith('/assets/') || /\.[a-f0-9]{8,}\.(js|css)$/i.test(pathname)) {
    return 'public, max-age=31536000, immutable';
  }

  if (ASSET_EXTENSIONS.test(pathname)) {
    return 'public, max-age=86400, s-maxage=604800';
  }

  return 'public, max-age=300, s-maxage=3600';
}

function stripFirstSegment(pathname) {
  const parts = pathname.split('/').filter(Boolean);
  return parts.length > 1 ? `/${parts.slice(1).join('/')}` : pathname;
}

let b2AuthCache = { token: null, downloadUrl: null, expiresAt: 0 };

async function ensureB2Auth(env) {
  const now = Date.now();
  if (b2AuthCache.token && now < b2AuthCache.expiresAt) return b2AuthCache;

  if (!env.B2_KEY_ID || !env.B2_APP_KEY) {
    throw new Error('missing-b2-credentials');
  }

  const res = await fetch('https://api.backblazeb2.com/b2api/v2/b2_authorize_account', {
    headers: {
      authorization: `Basic ${btoa(`${env.B2_KEY_ID}:${env.B2_APP_KEY}`)}`
    }
  });

  if (!res.ok) {
    const t = await res.text().catch(() => '');
    throw new Error(`b2_authorize_account failed (${res.status}): ${t}`);
  }
  const data = await res.json();
  b2AuthCache = {
    token: data.authorizationToken,
    downloadUrl: String(data.downloadUrl || '').replace(/\/$/, ''),
    expiresAt: now + 1000 * 60 * 60 * 12
  };
  return b2AuthCache;
}

async function fetchFromB2({ base, bucket, objectKey, acceptEncoding, authToken }) {
  const url = `${base}/file/${bucket}/${objectKey}`;
  const headers = new Headers();
  if (acceptEncoding) headers.set('accept-encoding', acceptEncoding);
  if (authToken) headers.set('authorization', authToken);
  return fetch(url, { headers });
}

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const hostname = (request.headers.get('x-forwarded-host') || url.hostname).toLowerCase();
    const { ROOT_DOMAIN, B2_DOWNLOAD_BASE, B2_BUCKET_NAME, B2_KEY_ID, B2_APP_KEY, ROUTING } = env;

    if (!B2_DOWNLOAD_BASE || !B2_BUCKET_NAME) {
      return new Response('Service misconfigured', { status: 500 });
    }

    let siteId = await ROUTING.get(`host:${hostname}`);
    if (!siteId) {
      const sub = parseSubdomain(hostname, ROOT_DOMAIN);
      if (sub) siteId = await ROUTING.get(`host:${sub}`);
    }
    if (!siteId) {
      return new Response('Site not found', { status: 404 });
    }

    const deployId = await ROUTING.get(`current:${siteId}`);
    if (!deployId) {
      return new Response('No deployment found', { status: 404 });
    }

    const pathname = url.pathname;
    const keyPath = pathname.replace(/^\//, '') || 'index.html';
    const basePath = `sites/${siteId}/${deployId}`;
    const acceptEncoding = request.headers.get('accept-encoding');

    let authToken = null;
    let base = (B2_DOWNLOAD_BASE || '').replace(/\/$/, '');

    if (B2_KEY_ID && B2_APP_KEY) {
      const auth = await ensureB2Auth(env);
      authToken = auth.token;
      base = auth.downloadUrl;
    }

    let res = await fetchFromB2({
      base,
      bucket: B2_BUCKET_NAME,
      objectKey: `${basePath}/${keyPath}`,
      acceptEncoding,
      authToken
    });

    if (res.status === 404 && isAssetPath(pathname)) {
      const rewritten = stripFirstSegment(pathname);
      if (rewritten !== pathname) {
        const rewrittenKey = rewritten.replace(/^\//, '');
        res = await fetchFromB2({
          base,
          bucket: B2_BUCKET_NAME,
          objectKey: `${basePath}/${rewrittenKey}`,
          acceptEncoding,
          authToken
        });
      }
    }

    if (res.status === 404 && !isAssetPath(pathname)) {
      const dirPath = pathname.endsWith('/') ? pathname : `${pathname}/`;
      const dirKey = `${keyPath.replace(/\/$/, '')}/index.html`;

      const dirRes = await fetchFromB2({
        base,
        bucket: B2_BUCKET_NAME,
        objectKey: `${basePath}/${dirKey}`,
        acceptEncoding,
        authToken
      });

      if (dirRes.ok) {
        res = dirRes;
      }
    }

    if (res.status === 404 && !isAssetPath(pathname)) {
      res = await fetchFromB2({
        base,
        bucket: B2_BUCKET_NAME,
        objectKey: `${basePath}/index.html`,
        acceptEncoding,
        authToken
      });
    }

    if (!res.ok) {
      return new Response('Not found', { status: 404 });
    }

    const headers = new Headers(res.headers);
    headers.set('cache-control', getCacheControl(pathname));
    headers.set('x-content-type-options', 'nosniff');

    headers.set('server', 'boop.cat');
    headers.set('x-boop-host', 'boop.cat');
    headers.set('x-boop-site-id', siteId);
    headers.set('x-boop-deploy-id', deployId);

    headers.delete('x-bz-file-id');
    headers.delete('x-bz-file-name');
    headers.delete('x-bz-content-sha1');
    headers.delete('x-bz-upload-timestamp');
    headers.delete('x-bz-info-src_last_modified_millis');

    return new Response(res.body, { status: 200, headers });
  }
};
