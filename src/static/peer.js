function createPeer() {
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
  })
  pc.onnegotiationneeded = handleNegotiaionNeeded
  pc.onicecandidate = handleIceCandidate
  pc.onconnectionstatechange = handleConnectionStateChange

  return pc
}

async function handleNegotiaionNeeded() {
  const offer = await pc.createOffer()
  await pc.setLocalDescription(offer)
  ws.send(
    JSON.stringify({
      type: "offer",
      payload: JSON.stringify(offer),
    })
  )
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

function handleConnectionStateChange(event) {
  console.log("Connection:", pc.connectionState)
  if (event.connectionState === "connected") {
    setStatus("Connected! Streaming video", "success")
  } else if (event.connectionState === "failed") {
    setStatus("Connection failed", "error")
  }
}
