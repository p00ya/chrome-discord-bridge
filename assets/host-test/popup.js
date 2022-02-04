var port = null;
const host = "io.github.p00ya.cdb_echo";

function log(s) {
  const output = document.getElementById("output");

  output.innerText = output.innerText + "\n" + s;
}

connect.addEventListener("click", () => {
  if (port !== null) {
    log("Already connected");
  }

  port = chrome.runtime.connectNative(host);
  log("Connected to " + host);

  port.onMessage.addListener((msg) => {
    log("Received: " + JSON.stringify(msg));
  });
  port.onDisconnect.addListener(() => {
    log("Disconnected by remote host");
    port = null;
  });
});

send.addEventListener("click", () => {
  if (port !== null) {
    port.postMessage({
      t: Date.now()
    });
  }
});

disconnect.addEventListener("click", () => {
  if (port === null) {
    return;
  }
  port.disconnect();
  port = null;
  log("Disconnected");
});
