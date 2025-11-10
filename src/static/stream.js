const video = document.getElementById("video")
let ws = new WebSocket(webSocketURL)

let pc
let statsInterval

ws.onopen = async () => {
  ws.send(
    JSON.stringify({
      streamId: id,
      type: "stream",
    })
  )
  console.log("Sent Stream")
  updateConnectionStatus("connecting")
}

ws.onmessage = async (event) => {
  const msg = JSON.parse(event.data)
  if (msg.type === "offer") {
    pc = createPeer(ws)
    pc.ontrack = (e) => {
      handleOnTrack(e, video)
      // Start collecting stats once track is received
      startStatsCollection()
    }

    pc.onconnectionstatechange = () => {
      console.log("Connection state:", pc.connectionState)
      updateConnectionStatus(pc.connectionState)

      if (
        pc.connectionState === "disconnected" ||
        pc.connectionState === "failed"
      ) {
        stopStatsCollection()
      }
    }

    const offer = JSON.parse(msg.payload)
    pc.setRemoteDescription(new RTCSessionDescription(offer))
    const answer = await pc.createAnswer()
    pc.setLocalDescription(answer)
    ws.send(
      JSON.stringify({
        type: "answer",
        isLiveStream: true,
        payload: JSON.stringify(answer),
      })
    )
  } else {
    handleWSEvent(msg, pc)
  }
}

ws.onerror = (error) => {
  console.error("WebSocket error:", error)
  updateConnectionStatus("failed")
}

ws.onclose = () => {
  console.log("WebSocket closed")
  updateConnectionStatus("closed")
  stopStatsCollection()
}

// Update connection status UI
function updateConnectionStatus(state) {
  const statusIndicator = document.getElementById("statusIndicator")
  const connectionStatus = document.getElementById("connectionStatus")

  if (!statusIndicator || !connectionStatus) return

  switch (state) {
    case "connected":
      statusIndicator.className = "status-indicator status-connected"
      connectionStatus.textContent = "Connected"
      break
    case "connecting":
    case "new":
      statusIndicator.className = "status-indicator status-connecting"
      connectionStatus.textContent = "Connecting..."
      break
    case "disconnected":
    case "failed":
    case "closed":
      statusIndicator.className = "status-indicator status-disconnected"
      connectionStatus.textContent = "Disconnected"
      break
    default:
      statusIndicator.className = "status-indicator status-connecting"
      connectionStatus.textContent = state
  }
}

// Start collecting WebRTC stats
function startStatsCollection() {
  if (statsInterval) {
    clearInterval(statsInterval)
  }

  let lastBytesReceived = 0
  let lastTimestamp = Date.now()

  statsInterval = setInterval(async () => {
    if (!pc) return

    try {
      const stats = await pc.getStats()
      let inboundStats = null
      let candidatePairStats = null

      stats.forEach((report) => {
        if (report.type === "inbound-rtp" && report.kind === "video") {
          inboundStats = report
        }
        if (report.type === "candidate-pair" && report.state === "succeeded") {
          candidatePairStats = report
        }
      })

      if (inboundStats) {
        const currentTime = Date.now()
        const currentBytes = inboundStats.bytesReceived || 0
        const timeDiff = (currentTime - lastTimestamp) / 1000 // in seconds
        const bytesDiff = currentBytes - lastBytesReceived

        // Calculate bitrate
        const bitrate = timeDiff > 0 ? (bytesDiff * 8) / timeDiff : 0

        // Update last values
        lastBytesReceived = currentBytes
        lastTimestamp = currentTime

        // Prepare stats object
        const statsData = {
          connected: pc.connectionState === "connected",
          bitrate: bitrate,
          packetsLost: inboundStats.packetsLost || 0,
          bytesReceived: inboundStats.bytesReceived || 0,
          framesReceived: inboundStats.framesReceived || 0,
          framesDropped: inboundStats.framesDropped || 0,
          frameRate: inboundStats.framesPerSecond || 0,
          resolution:
            inboundStats.frameWidth && inboundStats.frameHeight
              ? `${inboundStats.frameWidth}x${inboundStats.frameHeight}`
              : "-",
          codec: getCodecName(inboundStats.codecId, stats),
          jitter: inboundStats.jitter ? inboundStats.jitter * 1000 : 0, // Convert to ms
        }

        // Add candidate pair stats if available
        if (candidatePairStats) {
          statsData.rtt = candidatePairStats.currentRoundTripTime
            ? candidatePairStats.currentRoundTripTime * 1000
            : 0 // Convert to ms
          statsData.connectionType = getCandidateType(candidatePairStats, stats)
        }

        updateStats(statsData)
      }
    } catch (error) {
      console.error("Error collecting stats:", error)
    }
  }, 1000) // Update every second
}

// Stop collecting stats
function stopStatsCollection() {
  if (statsInterval) {
    clearInterval(statsInterval)
    statsInterval = null
  }
}

// Get codec name from codec ID
function getCodecName(codecId, stats) {
  if (!codecId) return "-"

  for (let report of stats.values()) {
    if (report.id === codecId && report.type === "codec") {
      return report.mimeType ? report.mimeType.split("/")[1].toUpperCase() : "-"
    }
  }
  return "-"
}

// Get connection type from candidate pair
function getCandidateType(candidatePair, stats) {
  if (!candidatePair.localCandidateId) return "Unknown"

  for (let report of stats.values()) {
    if (report.id === candidatePair.localCandidateId) {
      return report.protocol ? report.protocol.toUpperCase() : "Unknown"
    }
  }
  return "Unknown"
}

// Update UI with stats
function updateStats(stats) {
  // Update network stats
  if (stats.bitrate !== undefined) {
    const bitrateEl = document.getElementById("bitrate")
    if (bitrateEl)
      bitrateEl.textContent = `${(stats.bitrate / 1000).toFixed(2)} kbps`
  }

  if (stats.packetsLost !== undefined) {
    const packetsLostEl = document.getElementById("packetsLost")
    if (packetsLostEl) packetsLostEl.textContent = stats.packetsLost
  }

  if (stats.rtt !== undefined) {
    const rttEl = document.getElementById("rtt")
    if (rttEl) rttEl.textContent = `${stats.rtt.toFixed(0)} ms`
  }

  if (stats.jitter !== undefined) {
    const jitterEl = document.getElementById("jitter")
    if (jitterEl) jitterEl.textContent = `${stats.jitter.toFixed(2)} ms`
  }

  // Update transfer data
  if (stats.bytesReceived !== undefined) {
    const bytesReceivedEl = document.getElementById("bytesReceived")
    if (bytesReceivedEl)
      bytesReceivedEl.textContent = `${(
        stats.bytesReceived /
        1024 /
        1024
      ).toFixed(2)} MB`
  }

  if (stats.framesReceived !== undefined) {
    const framesReceivedEl = document.getElementById("framesReceived")
    if (framesReceivedEl) framesReceivedEl.textContent = stats.framesReceived
  }

  if (stats.framesDropped !== undefined) {
    const framesDroppedEl = document.getElementById("framesDropped")
    if (framesDroppedEl) framesDroppedEl.textContent = stats.framesDropped
  }

  // Update video quality
  if (stats.resolution) {
    const resolutionEl = document.getElementById("resolution")
    if (resolutionEl) resolutionEl.textContent = stats.resolution
  }

  if (stats.frameRate !== undefined) {
    const frameRateEl = document.getElementById("frameRate")
    if (frameRateEl) frameRateEl.textContent = `${stats.frameRate} fps`
  }

  if (stats.codec) {
    const codecEl = document.getElementById("codec")
    if (codecEl) codecEl.textContent = stats.codec
  }

  if (stats.connectionType) {
    const connectionTypeEl = document.getElementById("connectionType")
    if (connectionTypeEl) connectionTypeEl.textContent = stats.connectionType
  }
}
