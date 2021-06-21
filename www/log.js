"use strict";
function postJSON(url, data) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", url, true);
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.send(JSON.stringify(data));
}
var data = {
        ti: Math.round(Date.now() / 1000),
        tz: Intl.DateTimeFormat().resolvedOptions().timeZone,
        ua: window.navigator.userAgent,
        re: document.referrer,
        ho: location.host,
        pa: location.pathname,
};
console.log(JSON.stringify(data));
postJSON("https://s0rvs5.deta.dev/", data);
