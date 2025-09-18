export async function onRequest(context) {
  const { request, env } = context;
  const url = new URL(request.url);
  const assetResponse = await env.ASSETS.fetch(request);

  if (!assetResponse || assetResponse.status === 404) {
    return assetResponse || new Response('Not found', { status: 404 });
  }

  const headers = new Headers(assetResponse.headers);

  if (url.pathname.endsWith('.pagefind')) {
    headers.set('Content-Type', 'application/wasm');
    headers.set('Cache-Control', 'public, max-age=31536000, immutable');
  } else if (url.pathname.endsWith('.pf_index') || url.pathname.endsWith('.pf_fragment') || url.pathname.endsWith('.pf_meta') || url.pathname.endsWith('.pf_filter')) {
    headers.set('Content-Type', 'application/octet-stream');
    headers.set('Cache-Control', 'public, max-age=604800');
  } else if (url.pathname.endsWith('.js')) {
    headers.set('Content-Type', 'application/javascript; charset=utf-8');
    headers.set('Cache-Control', 'public, max-age=31536000, immutable');
  } else if (url.pathname.endsWith('.css')) {
    headers.set('Content-Type', 'text/css; charset=utf-8');
    headers.set('Cache-Control', 'public, max-age=31536000, immutable');
  }

  return new Response(assetResponse.body, { status: assetResponse.status, headers });
}
