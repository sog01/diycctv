const video = document.getElementById("video")
const status = document.getElementById("status")
const startBtn = document.getElementById("startBtn")
const stopBtn = document.getElementById("stopBtn")

let ws, pc, stream

function setStatus(msg, type = "info") {
  status.textContent = msg
  status.className = type
  console.log(msg)
}

async function start() {
  try {
    setStatus("Connecting to server...", "info")

    ws = new WebSocket(webSocketURL)
    ws.onopen = async () => {
      setStatus("Getting camera...", "info")

      stream = await navigator.mediaDevices.getUserMedia({
        video: true,
        audio: false,
      })

      video.srcObject = stream

      setStatus("Creating peer connection...", "info")
      pc = createPeer()
      await stream.getTracks().forEach((track) => {
        pc.addTrack(track, stream)
      })
      setStatus("Sent offer, waiting for answer...", "info")
    }

    ws.onmessage = async (event) => {
      const msg = JSON.parse(event.data)
      handleWSEvent(msg, pc)
    }

    ws.onerror = (error) => {
      console.error("WebSocket error:", error)
      setStatus("WebSocket error", "error")
    }

    ws.onclose = () => {
      console.log("WebSocket closed")
      setStatus("Connection closed", "error")
    }

    startBtn.disabled = true
    stopBtn.disabled = false
  } catch (error) {
    console.error("Error:", error)
    setStatus("Error: " + error.message, "error")
    stop()
  }
}

function stop() {
  if (stream) {
    stream.getTracks().forEach((t) => t.stop())
    stream = null
  }
  if (pc) {
    pc.close()
    pc = null
  }
  if (ws) {
    ws.close()
    ws = null
  }
  video.srcObject = null
  setStatus("Stopped", "info")
  startBtn.disabled = false
  stopBtn.disabled = true
}

window.addEventListener("beforeunload", stop)
