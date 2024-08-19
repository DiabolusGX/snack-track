chrome.runtime.onInstalled.addListener(() => {
    console.log("Snack track extension installed");
    chrome.action.setBadgeText({
        text: "OFF",
    });
});

chrome.action.onClicked.addListener(async (tab) => {
    console.log(`Action clicked on tab: `, tab);

    chrome.tabs.sendMessage(tab.id, { type: "auth-status" }, async (response) => {
        await chrome.action.setBadgeText({
            tabId: tab.id,
            text: response.isLoggedIn ? "ON" : "OFF",
        });
    });
});
