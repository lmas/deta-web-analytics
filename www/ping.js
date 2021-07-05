"use strict";

// WIP: Super simple and privacy friendly analytics.

// If this script ends up somewhere else.. try and avoid running at all.
if (location.host != "www.larus.se") throw new Error("invalid host");

function postJSON(url, data) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", url, true);
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.send(JSON.stringify(data));
}

postJSON("https://stats.larus.se/ping", {
        tz: Intl.DateTimeFormat().resolvedOptions().timeZone,
        ua: window.navigator.userAgent,
        re: document.referrer,
        ho: location.host,
        pa: location.pathname,
});

