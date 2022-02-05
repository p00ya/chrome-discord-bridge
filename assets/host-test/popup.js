let port = null;

// Discord app ID for chrome-discord-bridge dev.
const exampleAppId = "922040684020645908";

const host = document.getElementById("host");
const connect = document.getElementById("connect");
const handshake = document.getElementById("handshake");
const setactivity = document.getElementById("setactivity");
const disconnect = document.getElementById("disconnect");
const output = document.getElementById("output");

function log(s) {
  output.innerText = output.innerText + "\n" + s;
}

function reset() {
  port = null;
  host.disabled = false;
  connect.disabled = false;
  disconnect.disabled = true;
}

connect.addEventListener("click", () => {
  if (port !== null) {
    log("Already connected");
  }

  port = chrome.runtime.connectNative(host.value);
  log("Connected to " + host.value);
  host.disabled = true;
  connect.disabled = true;
  disconnect.disabled = false;

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
