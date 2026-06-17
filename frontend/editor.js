import { EditorState } from "https://esm.sh/@codemirror/state"
import { EditorView, keymap, lineNumbers, highlightActiveLine } from "https://esm.sh/@codemirror/view"
import { defaultKeymap, history, historyKeymap } from "https://esm.sh/@codemirror/commands"
import { syntaxHighlighting, defaultHighlightStyle } from "https://esm.sh/@codemirror/language"
import { collab, getSyncedVersion, receiveUpdates, sendableUpdates, getClientID } from "https://esm.sh/@codemirror/collab"
import { javascript } from "https://esm.sh/@codemirror/lang-javascript"
import { python } from "https://esm.sh/@codemirror/lang-python"
import { go } from "https://esm.sh/@codemirror/lang-go"
import { java } from "https://esm.sh/@codemirror/lang-java"
import { cpp } from "https://esm.sh/@codemirror/lang-cpp"
import { rust } from "https://esm.sh/@codemirror/lang-rust"
import { html } from "https://esm.sh/@codemirror/lang-html"
import { css } from "https://esm.sh/@codemirror/lang-css"
import { json } from "https://esm.sh/@codemirror/lang-json"
import { oneDark } from "https://esm.sh/@codemirror/theme-one-dark"

// --- URL params ---
const params = new URLSearchParams(window.location.search)
const roomId = params.get('room')
const username = params.get('username')

if (!roomId || !username) {
  window.location.href = '/'
}

// --- Room ID display + copy ---
const roomIdDisplay = document.getElementById('room-id-display')
const copyHint = document.getElementById('copy-hint')
roomIdDisplay.textContent = roomId
roomIdDisplay.addEventListener('click', () => {
  navigator.clipboard.writeText(roomId)
  copyHint.textContent = 'copied!'
  setTimeout(() => copyHint.textContent = '', 2000)
})

// --- Language map ---
const langMap = {
  javascript: javascript(),
  typescript: javascript({ typescript: true }),
  python: python(),
  go: go(),
  java: java(),
  cpp: cpp(),
  rust: rust(),
  html: html(),
  css: css(),
  json: json(),
}

// --- Language compartment for dynamic switching ---
import { Compartment } from "https://esm.sh/@codemirror/state"
const languageCompartment = new Compartment()

// --- WebSocket ---
const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
const ws = new WebSocket(`${wsProtocol}://${window.location.host}/ws/${roomId}?username=${encodeURIComponent(username)}`)

// --- Editor setup (after sync received) ---
let view = null

function buildEditor(doc, version, language) {
  const state = EditorState.create({
    doc,
    extensions: [
      history(),
      lineNumbers(),
      highlightActiveLine(),
      syntaxHighlighting(defaultHighlightStyle),
      keymap.of([...defaultKeymap, ...historyKeymap]),
      languageCompartment.of(langMap[language] || langMap.javascript),
      oneDark,
      collab({ startVersion: version }),
      EditorView.updateListener.of(update => {
        if (update.docChanged) {
          pushUpdates(view)
        }
      }),
    ]
  })

  view = new EditorView({
    state,
    parent: document.getElementById('editor')
  })
}

// --- Push local updates to server ---
function pushUpdates(view) {
  const updates = sendableUpdates(view.state)
  if (updates.length === 0) return

  const version = getSyncedVersion(view.state)

  ws.send(JSON.stringify({
    type: 'update',
    payload: {
      clientId: getClientID(view.state),
      version,
      changes: updates.map(u => ({
        changes: u.changes.toJSON(),
        clientID: u.clientID
      }))
    }
  }))
}

// --- Apply remote updates to editor ---
function applyUpdates(view, updates) {
  const mapped = updates.map(u => ({
    changes: view.state.changes(u.changes),
    clientID: u.clientID
  }))
  view.dispatch(receiveUpdates(view.state, mapped))
}

// --- WebSocket message handler ---
ws.addEventListener('message', (event) => {
  const { type, payload } = JSON.parse(event.data)

  switch (type) {
    case 'sync': {
      buildEditor(payload.doc, payload.version, payload.language)
      document.getElementById('lang-select').value = payload.language
      updatePresence(payload.users)
      break
    }

    case 'update': {
      if (!view) return
      applyUpdates(view, payload.changes)
      break
    }

    case 'presence': {
      updatePresence(payload.users)
      break
    }

    case 'lang_change': {
      if (!view) return
      document.getElementById('lang-select').value = payload.language
      view.dispatch({
        effects: languageCompartment.reconfigure(langMap[payload.language] || langMap.javascript)
      })
      break
    }
  }
})

ws.addEventListener('close', () => {
  console.log('disconnected from room')
})

// --- Language select handler ---
document.getElementById('lang-select').addEventListener('change', (e) => {
  const language = e.target.value

  if (view) {
    view.dispatch({
      effects: languageCompartment.reconfigure(langMap[language] || langMap.javascript)
    })
  }

  ws.send(JSON.stringify({
    type: 'lang_change',
    payload: { language }
  }))
})

// --- Presence ---
function updatePresence(users) {
  const presenceEl = document.getElementById('presence-list')
  if (!users || users.length === 0) {
    presenceEl.textContent = 'just you'
    return
  }
  presenceEl.textContent = users.join(', ')
}