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
    pc.ontrack = (e) => {
      handleOnTrack(e, video)
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
}

ws.onclose = () => {
  console.log("WebSocket closed")
}
