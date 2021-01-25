addEventListener('fetch', event => {
    event.respondWith(handleRequest(event.request))
})

const DETAILS_URL = "https://raw.githubusercontent.com/hellodword/wechat-feeds/feeds/details.json";
const DEFAULT_ICO = "https://wechat.privacyhide.com/favicon.ico";

/**
 * Respond to the request
 * @param {Request} request
 */
async function handleRequest(request) {
    try {
        let u = new URL(request.url);

        let details = await fetch(DETAILS_URL).then((j) => {
            return j.json();
        });

        let bizid = u.searchParams.get("host").split(".")[0];
        for (let k of details) {
            if ((k.bizid+'').replace(/=/g, "").toLowerCase() == bizid.toLowerCase()){
                return fetch((k.head_img+"").replace("/132", "/64"));
              }
        }
        return fetch(DEFAULT_ICO);
    } catch (e) {
        return fetch(DEFAULT_ICO);
    }
}