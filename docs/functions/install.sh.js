// Cloudflare Pages Function: Serve /install.sh by proxying the GitHub raw file.
// Keeps https://devbox.ar0.eu/install.sh in sync with the repository's install.sh.

export async function onRequest(context) {
  const upstreamUrl = 'https://devbox.ar0.eu/install.sh';

  // Forward conditional request headers for better caching (ETag support)
  const headers = new Headers();
  const ifNoneMatch = context.request.headers.get('If-None-Match');
  const ifModifiedSince = context.request.headers.get('If-Modified-Since');
  if (ifNoneMatch) headers.set('If-None-Match', ifNoneMatch);
  if (ifModifiedSince) headers.set('If-Modified-Since', ifModifiedSince);

  try {
    const upstream = await fetch(upstreamUrl, {
      headers,
      // Hint Cloudflare cache; even if ignored, upstream ETag will help
      cf: { cacheEverything: true, cacheTtl: 300 },
    });

    if (upstream.status === 304) {
      return new Response(null, { status: 304 });
    }

    if (upstream.ok) {
      const respHeaders = new Headers(upstream.headers);
      // Ensure correct content type
      respHeaders.set('Content-Type', 'text/x-shellscript; charset=utf-8');
      // Cache for 5 minutes at edge/browsers
      respHeaders.set('Cache-Control', 'public, max-age=300, s-maxage=300');
      // Optional CORS for curl/wget from other origins
      respHeaders.set('Access-Control-Allow-Origin', '*');
      return new Response(upstream.body, { status: 200, headers: respHeaders });
    }

    // Fallback: temporary redirect to the upstream if proxy fetch fails
    return Response.redirect(upstreamUrl, 302);
  } catch (err) {
    return Response.redirect(upstreamUrl, 302);
  }
}
