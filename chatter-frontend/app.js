const API = "http://localhost:8080";
const WS = "ws://localhost:8080/api/v1/ws";

const state = {
  accessToken: null,
  userId: null,
  ws: null,
  conversation: null,
  conversations: [],
  messages: new Map(),
  lastCursor: null
};

const $ = (id) => document.getElementById(id);

function logEvent(text) {
  const time = new Date().toLocaleTimeString();
  $("eventLog").textContent = `[${time}] ${text}\n` + $("eventLog").textContent;
}

function setStatus(text, ok = false) {
  $("authStatus").textContent = text;
  $("authStatus").className = `status ${ok ? "success" : "error"}`;
}

function api(path, options = {}) {
  return fetch(`${API}${path}`, {
    ...options,
    headers: {
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...(state.accessToken ? { Authorization: `Bearer ${state.accessToken}` } : {}),
      ...(options.headers || {})
    }
  });
}

function showChat(userId) {
  state.userId = userId;
  $("authView").classList.add("hidden");
  $("chatView").classList.remove("hidden");
  $("currentUser").textContent = userId;
  $("debugStatus").textContent = "Authenticated";
  loadConversations();
}

function formatTime(value) {
  return new Date(value).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

function renderMessages() {
  const container = $("messages");
  if (!state.messages.size) {
    container.innerHTML = `<div class="empty"><div class="empty-icon">↗</div><h3>No messages</h3><p>Send a message or load history.</p></div>`;
    return;
  }

  const messages = [...state.messages.values()].sort((a, b) =>
    new Date(a.created_at) - new Date(b.created_at) || a.id.localeCompare(b.id)
  );

  container.innerHTML = "";
  for (const msg of messages) {
    const mine = msg.sender_id === state.userId;
    const el = document.createElement("article");
    el.className = `message ${mine ? "mine" : ""}`;
    el.innerHTML = `
      <div class="meta">${mine ? "You" : msg.sender_id}</div>
      <div class="body"></div>
      <div class="time">${formatTime(msg.created_at)}</div>`;
    el.querySelector(".body").textContent = msg.content;
    container.appendChild(el);
  }
  container.scrollTop = container.scrollHeight;
}

function addMessages(messages) {
  for (const msg of messages || []) state.messages.set(msg.id, msg);
  renderMessages();
}

function selectConversation(conv) {
  state.conversation = conv;
  state.messages.clear();
  $("conversationTitle").textContent = conv.other_username || conv.other_user_id || "Conversation";
  $("conversationId").textContent = conv.id;
  $("loadHistoryBtn").disabled = false;
  $("syncBtn").disabled = !state.ws || state.ws.readyState !== WebSocket.OPEN;
  $("messageInput").disabled = !state.ws || state.ws.readyState !== WebSocket.OPEN;
  $("messageForm").querySelector("button").disabled = !state.ws || state.ws.readyState !== WebSocket.OPEN;
  renderConversationList();
  renderMessages();
}

function renderConversationList() {
  const list = $("conversationList");
  list.innerHTML = "";
  for (const conv of state.conversations) {
    const button = document.createElement("button");
    button.className = `conversation ${state.conversation?.id === conv.id ? "active" : ""}`;
    button.innerHTML = `<div class="name"></div><div class="id"></div>`;
    button.querySelector(".name").textContent = conv.other_username || conv.other_user_id;
    button.querySelector(".id").textContent = conv.id;
    button.addEventListener("click", () => selectConversation(conv));
    list.appendChild(button);
  }
  if (!state.conversations.length) {
    list.innerHTML = `<div class="status">No conversations found.</div>`;
  }
}

async function loadConversations() {
  try {
    const res = await api("/api/v1/conversations");
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    state.conversations = await res.json();
    if (Array.isArray(state.conversations) === false && state.conversations.conversations) {
      state.conversations = state.conversations.conversations;
    }
    renderConversationList();
  } catch (err) {
    logEvent(`conversation list failed: ${err.message}`);
  }
}

async function loadHistory() {
  if (!state.conversation) return;

  try {
    const res = await api(
      `/api/v1/conversations/${state.conversation.id}/messages`
    );

    if (!res.ok) throw new Error(`HTTP ${res.status}`);

    const data = await res.json();

    addMessages(data.messages || []);

    state.lastCursor = null;

    $("debugCursor").textContent = "—";

    logEvent(
      `history loaded: ${(data.messages || []).length} messages`
    );
  } catch (err) {
    logEvent(`history failed: ${err.message}`);
  }
}

function connectWebSocket() {
  if (!state.accessToken) return;
  if (state.ws && state.ws.readyState === WebSocket.OPEN) return;

  const ws = new WebSocket(
    `${WS}?token=${encodeURIComponent(state.accessToken)}`
  );
  state.ws = ws;
  $("connectionText").textContent = "Connecting…";
  $("debugSocket").textContent = "Connecting";

  ws.onopen = () => {
    $("connectionDot").classList.add("online");
    $("connectionText").textContent = "Connected";
    $("debugSocket").textContent = "Open";
    $("debugStatus").textContent = "Live";
    $("connectBtn").textContent = "Disconnect";
    $("messageInput").disabled = !state.conversation;
    $("messageForm").querySelector("button").disabled = !state.conversation;
    $("syncBtn").disabled = !state.conversation;
    logEvent("WebSocket connected");
  };

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);

      logEvent(`← ${data.type}`);

      if (data.type === "message" && data.message) {
        const msg = data.message;

        // Only display messages for the currently selected conversation.
        if (
          state.conversation &&
          msg.conversation_id === state.conversation.id
        ) {
          state.messages.set(msg.id, msg);
          renderMessages();
        }

        updateCursorFromMessage(msg);
      }

      if (data.type === "sync") {
        const messages = data.messages || [];

        addMessages(messages);

        if (messages.length > 0) {
          updateCursorFromMessage(messages[messages.length - 1]);
        }

        $("eventBanner").textContent =
          `Synced ${messages.length} message(s)`;

        $("eventBanner").classList.remove("hidden");
      }

      if (data.type === "error") {
        logEvent(`server error: ${data.error}`);
      }
    } catch (err) {
      logEvent(`WebSocket message error: ${err.message}`);
    }
  };

  ws.onerror = () => logEvent("WebSocket error");

  ws.onclose = () => {
    $("connectionDot").classList.remove("online");
    $("connectionText").textContent = "Disconnected";
    $("debugSocket").textContent = "Closed";
    $("debugStatus").textContent = "Offline";
    $("connectBtn").textContent = "Connect";
    $("messageInput").disabled = true;
    $("messageForm").querySelector("button").disabled = true;
    $("syncBtn").disabled = true;
    logEvent("WebSocket disconnected");
  };
}

function disconnectWebSocket() {
  if (state.ws) {
    state.ws.close();
    state.ws = null;
  }
}

function updateCursorFromMessage(msg) {
  $("debugCursor").textContent = `${msg.created_at} / ${msg.id}`;
}

function sendMessage(content) {
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN || !state.conversation) return;
  state.ws.send(JSON.stringify({
    type: "message",
    conversation_id: state.conversation.id,
    content
  }));
  logEvent("→ message");
}

function syncMessages() {
  if (!state.ws || state.ws.readyState !== WebSocket.OPEN || !state.conversation) return;
  if (!state.lastCursor) {
    logEvent("No encoded history cursor available. Load history first.");
    return;
  }
  state.ws.send(JSON.stringify({
    type: "sync",
    conversation_id: state.conversation.id,
    after: state.lastCursor
  }));
  logEvent("→ sync");
}

async function login(email, password) {
  const res = await fetch(`${API}/api/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password })
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  state.accessToken = data.access_token;
  // The API does not return user ID directly, so decode JWT payload locally.
  const payload = JSON.parse(atob(data.access_token.split(".")[1].replace(/-/g, "+").replace(/_/g, "/")));
  showChat(payload.sub);
}

async function register(username, email, password) {
  const res = await fetch(`${API}/api/v1/users`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, email, password })
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  setStatus("Account created. You can now log in.", true);
}

document.querySelectorAll(".tab").forEach(tab => {
  tab.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach(t => t.classList.remove("active"));
    tab.classList.add("active");
    const loginMode = tab.dataset.auth === "login";
    $("loginForm").classList.toggle("hidden", !loginMode);
    $("registerForm").classList.toggle("hidden", loginMode);
    $("authStatus").textContent = "";
  });
});

$("loginForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    setStatus("Logging in…", true);
    await login($("loginEmail").value.trim(), $("loginPassword").value);
    $("authStatus").textContent = "";
    connectWebSocket();
  } catch (err) {
    setStatus(err.message);
  }
});

$("registerForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    await register($("registerUsername").value.trim(), $("registerEmail").value.trim(), $("registerPassword").value);
  } catch (err) {
    setStatus(err.message);
  }
});

$("connectBtn").addEventListener("click", () => {
  if (state.ws?.readyState === WebSocket.OPEN) disconnectWebSocket();
  else connectWebSocket();
});

$("logoutBtn").addEventListener("click", () => {
  disconnectWebSocket();
  state.accessToken = null;
  state.userId = null;
  state.conversation = null;
  state.conversations = [];
  state.messages.clear();
  $("chatView").classList.add("hidden");
  $("authView").classList.remove("hidden");
});

$("refreshConversations").addEventListener("click", loadConversations);
$("loadHistoryBtn").addEventListener("click", loadHistory);
$("syncBtn").addEventListener("click", syncMessages);

$("messageForm").addEventListener("submit", (e) => {
  e.preventDefault();
  const input = $("messageInput");
  const content = input.value.trim();
  if (!content) return;
  sendMessage(content);
  input.value = "";
  input.focus();
});

$("clearLog").addEventListener("click", () => $("eventLog").textContent = "");

$("createConversationBtn").addEventListener("click", async () => {
  const userId = $("otherUserId").value.trim();
  if (!userId) return;
  try {
    const res = await api("/api/v1/conversations", {
      method: "POST",
      body: JSON.stringify({ user_id: userId })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
    await loadConversations();
    const id = data.id || data.conversation_id;
    const found = state.conversations.find(c => c.id === id);
    if (found) selectConversation(found);
    logEvent(`conversation opened: ${id}`);
  } catch (err) {
    logEvent(`create conversation failed: ${err.message}`);
  }
});
