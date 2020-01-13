let token

self.addEventListener('install', function (event) {
    event.waitUntil(self.skipWaiting()); // Activate worker immediately
});

self.addEventListener('activate', function (event) {
    event.waitUntil(self.clients.claim()); // Become available to all pages
});

self.addEventListener('message', ev => {
    token = ev.data
})

self.addEventListener("fetch", ev => {
    if (ev.request.url === "http://localhost:8080/user-profile") {
        const option = {headers: { ...ev.request.headers } }
        option.headers["Authorization"] = token
        ev.respondWith(fetch(ev.request.url, option))
    }
})