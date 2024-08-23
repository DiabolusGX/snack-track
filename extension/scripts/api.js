export default {
    getAddresses,
    fetchOrders,
    callWebhook,

    orderUpdateEndpoint: "order-update",
    userSettingsEndpoint: "user-settings",
};

const webhookHost = "http://snack-track.diabolus.me/webhook/";

async function fetchOrders() {
    const zh = await getZomatoHeaders();
    const requestOptions = {
        method: 'GET',
        headers: zh,
        redirect: 'follow'
    };

    const page = 0; // pagination not needed
    const response = await fetchWrrapper(`https://www.zomato.com/webroutes/user/orders?page=${page}`, requestOptions);
    if (response === "") {
        return {};
    }

    const ordersData = JSON.parse(response)?.entities?.ORDER;
    if (JSON.stringify(ordersData) === JSON.stringify([])) {
        return [];
    }

    const orderList = Object.keys(ordersData).map(orderId => ({
        id: orderId,
        ...ordersData[orderId]
    }));
    return orderList;
}

async function getAddresses() {
    // TODO: make actual API request
    return [
        {
            "id": 123,
            "address": "test address",
            "display_title": "Work",
            "display_subtitle": "here is work address",
        }
    ];
}

async function callWebhook(endpoint, payload) {
    const payloadStr = JSON.stringify(payload);

    const requestOptions = {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: payloadStr,
    };

    return fetchWrrapper(webhookHost + endpoint, requestOptions);
}

/******************** UTIL ********************/

async function fetchWrrapper(url, options) {
    return fetch(url, options).then(response => {
        return response.text();
    }).catch(_ => {
        return "";
    });
}

async function getZomatoHeaders() {
    const cookies = await new Promise((resolve, reject) => {
        chrome.cookies.getAll({ url: "https://www.zomato.com" }, (cookies) => {
            if (chrome.runtime.lastError) {
                return reject(chrome.runtime.lastError);
            }
            resolve(cookies);
        });
    });

    const cookieMap = {};
    cookies.forEach(cookie => {
        cookieMap[cookie.name] = cookie.value;
    });
    const cookieString = Object.keys(cookieMap).map(key => `${key}=${cookieMap[key]}`).join('; ');

    const zh = new Headers();
    zh.append("authority", "www.zomato.com");
    zh.append("sec-ch-ua", "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"99\", \"Google Chrome\";v=\"99\"");
    zh.append("x-zomato-csrft", cookieMap?.csrf);
    zh.append("sec-ch-ua-mobile", "?1");
    zh.append("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.82 Mobile Safari/537.36");
    zh.append("sec-ch-ua-platform", "\"Android\"");
    zh.append("accept", "*/*");
    zh.append("sec-fetch-site", "same-origin");
    zh.append("sec-fetch-mode", "cors");
    zh.append("sec-fetch-dest", "empty");
    zh.append("accept-language", "en-US,en;q=0.9");
    zh.append("cookie", cookieString);
    return zh;
}
