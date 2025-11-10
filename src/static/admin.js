const videos = document.querySelectorAll("video")
videos.forEach((video) => {
  let ws = new WebSocket(webSocketURL)
  ws.onopen = async () => {
    const streamId = video.getAttribute("streamId")
    ws.send(
      JSON.stringify({
        streamId: streamId,
        type: "stream",
      })
    )
    console.log("Sent Stream", streamId)
  }

  let pc
  ws.onmessage = async (event) => {
    const msg = JSON.parse(event.data)
    if (msg.type === "offer") {
      pc = createPeer(ws)
      const video = document.querySelector(`video[streamId="${msg.streamId}"]`)
      console.log(video)
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
})
