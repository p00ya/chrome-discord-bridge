var port = null;

// Discord app ID for chrome-discord-bridge dev.
const exampleAppId = "922040684020645908";

function log(s) {
  const output = document.getElementById("output");

  output.innerText = output.innerText + "\n" + s;
}

function reset() {
  port = null;
  document.getElementById("host").disabled = false;
  document.getElementById("connect").disabled = false;
  document.getElementById("disconnect").disabled = true;
}

connect.addEventListener("click", () => {
  if (port !== null) {
    log("Already connected");
  }

  const host = document.getElementById("host").value;

  port = chrome.runtime.connectNative(host);
  log("Connected to " + host);
  document.getElementById("host").disabled = true;
  document.getElementById("connect").disabled = true;
  document.getElementById("disconnect").disabled = false;

  port.onMessage.addListener((msg) => {
    log("Received: " + JSON.stringify(msg));
  });
  port.onDisconnect.addListener(() => {
    log("Disconnected by remote host");
    reset();
  });
});

handshake.addEventListener("click", () => {
  if (port === null) {
    return;
  }

  port.postMessage({
    v: 1,
    client_id: exampleAppId,
    nonce: Date.now()
  });
});

setactivity.addEventListener("click", () => {
  if (port === null) {
    return;
  }

  port.postMessage({
    nonce: Date.now(),
    cmd: "SET_ACTIVITY",
    args: {
      pid: 0, // not really our PID, but Discord complains if missing
      activity: {
        state: "Testing with host-test"
      }
    }
  });
});

disconnect.addEventListener("click", () => {
  if (port === null) {
    return;
  }
  port.disconnect();
  log("Disconnected");
  reset();
});
