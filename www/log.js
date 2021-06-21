"use strict";
function postJSON(url, data) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", url, true);
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.send(JSON.stringify(data));
}
var data = {
        ts: Math.round(Date.now() / 1000),
        tz: Intl.DateTimeFormat().resolvedOptions().timeZone,
        ua: window.navigator.userAgent,
        ref: document.referrer,
        url: [location.protocol, '//', location.host, location.pathname].join(''),
};
console.log(JSON.stringify(data));
postJSON("https://s0rvs5.deta.dev/", data);
