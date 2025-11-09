function createPeer() {
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
  })
  pc.onnegotiationneeded = sendOffer
  pc.onicecandidate = handleIceCandidate
  pc.onconnectionstatechange = (event) => {
    handleConnectionStateChange(event, pc)
  }

  return pc
}

async function sendOffer() {
  const offer = await pc.createOffer()
  await pc.setLocalDescription(offer)
  ws.send(
    JSON.stringify({
      type: "offer",
      payload: JSON.stringify(offer),
    })
  )
  console.log("Sent offer")
}

function handleIceCandidate(event) {
  if (event.candidate) {
    ws.send(
      JSON.stringify({
        type: "candidate",
        payload: JSON.stringify(event.candidate.toJSON()),
      })
    )
    console.log("Sent ICE candidate")
  }
}

function handleConnectionStateChange(event, pc) {
  console.log("Connection:", pc.connectionState)
  if (event.connectionState === "connected") {
    setStatus("Connected! Streaming video", "success")
  } else if (event.connectionState === "failed") {
    setStatus("Connection failed", "error")
  }
}

async function handleOnTrack(e) {
  console.log("Received Tracks", e)
  console.log("Track kind:", e.track.kind)
  console.log("Streams:", e.streams)
  console.log("Stream count:", e.streams.length)

  if (!e.streams || e.streams.length === 0) {
    console.error("No streams in track event!")
    return
  }

  const stream = e.streams[0]
  console.log("Stream ID:", stream.id)
  console.log("Stream tracks:", stream.getTracks())

  video.srcObject = stream
  console.log("Video srcObject set")

  try {
    await video.play()
    console.log("Video is playing!")
  } catch (err) {
    console.error("Video play error:", err)
  }
}

async function handleOnTrack(e, video) {
  console.log("Received Tracks", e)
  console.log("Track kind:", e.track.kind)
  console.log("Streams:", e.streams)
  console.log("Stream count:", e.streams.length)

  if (!e.streams || e.streams.length === 0) {
    console.error("No streams in track event!")
    return
  }

  const stream = e.streams[0]
  console.log("Stream ID:", stream.id)
  console.log("Stream tracks:", stream.getTracks())

  video.srcObject = stream
  console.log("Video srcObject set")

  // try {
  //   await video.play()
  //   console.log("Video is playing!")
  // } catch (err) {
  //   console.error("Video play error:", err)
  // }
}

async function handleWSEvent(msg, pc) {
  if (msg.type === "answer") {
    console.log("Received answer")
    const answer = JSON.parse(msg.payload)
    await pc.setRemoteDescription(answer)
  } else if (msg.type === "candidate") {
    console.log("Received ICE candidate")
    const candidate = JSON.parse(msg.payload)
    await pc.addIceCandidate(candidate)
  }
}
