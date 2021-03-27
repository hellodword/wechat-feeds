addEventListener('fetch', (event) => {
  event.respondWith(handleRequest(event.request))
})


function cloneHeaders(headers: Headers): Headers{
  let clone: Headers = new (Headers);

  for (let key of Array.from(headers.keys())) {
    clone.set(key, headers.get(key) || '')
  }

  return clone
}

export async function handleRequest(request: Request): Promise<Response> {
  let url = new URL(request.url);
  let true_url = request.url.replace(url.protocol + '//' + url.host + '/', '');
  true_url = true_url.replace(/^\/+/, '');
  if (!true_url.startsWith('http://') && !true_url.startsWith('https://')) {
    true_url = url.protocol + '//' + true_url;
  }

  if (true_url.indexOf('https://github.com/hellodword/wechat-feeds/raw/feeds/') !== 0 && true_url.indexOf('https://raw.githubusercontent.com/hellodword/wechat-feeds/feeds/') !== 0 )   {
    throw('go away')
  }


  let res = await fetch(true_url,
    {
      method: request.method,
      body: request.body,
      headers: request.headers,
      keepalive: request.keepalive,
    }
  )

  let headers = cloneHeaders(res.headers)

  headers.set('content-type', 'application/xml')

  return new Response(res.body, {
    headers: headers,
    status: res.status,
    statusText: res.statusText,
  })
}