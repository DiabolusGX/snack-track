console.log("Snack track extension content script loaded");

chrome.runtime.onMessage.addListener(function (msg, sender, sendResponse) {
    if (msg.type === "auth-status") {
        const loginSearchText = "Log in";
        const signupSearchText = "Sign up";

        const aTagsArray = Array.from(document.getElementsByTagName("a"));
        const loginTag = aTagsArray.filter(tag => tag.innerText.includes(loginSearchText));
        const signupTag = aTagsArray.filter(tag => tag.innerText.includes(signupSearchText));
        sendResponse({ isLoggedIn: !loginTag.length && !signupTag.length });
    }
});
