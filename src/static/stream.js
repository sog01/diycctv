const video = document.getElementById("video")
let ws = new WebSocket(webSocketURL)

let pc

ws.onopen = async () => {
  ws.send(
    JSON.stringify({
      streamId: id,
      type: "stream",
    })
  )
  console.log("Sent Stream")
}

ws.onmessage = async (event) => {
  const msg = JSON.parse(event.data)
  if (msg.type === "offer") {
    pc = createPeer()
    pc.ontrack = handleOnTrack
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
  } else if (msg.type === "answer") {
    console.log("Received answer")
    const answer = JSON.parse(msg.payload)
    await pc.setRemoteDescription(answer)
  } else if (msg.type === "candidate") {
    console.log("Received ICE candidate")
    const candidate = JSON.parse(msg.payload)
    await pc.addIceCandidate(candidate)
  }
}

async function handleOnTrack(e) {
  console.log("Received Tracks", e)
  console.log("Track kind:", e.track.kind)
  console.log("Streams:", e.streams)
  console.log("Stream count:", e.streams.length)

  // ⭐ MAKE SURE WE HAVE A STREAM
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

ws.onerror = (error) => {
  console.error("WebSocket error:", error)
}

ws.onclose = () => {
  console.log("WebSocket closed")
}
