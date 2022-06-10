<script setup>
import ConversationList from "./components/ConversationList.vue";
import MessageBoard from "./components/MessageBoard.vue";
</script>

<template>
  <div class="container mx-auto">
    <div class="min-w-full border rounded grid grid-cols-3">
      <ConversationList :id="id" :conversationsList="conversationsList"
        :onConversationCardClick="onConversationCardClick" :refreshConversationList="refreshConversationList"
        :createGroup="createGroup" />
      <MessageBoard :id="id" :otherConversation="otherConversation" :messages="otherConversationMessages"
        :sendMessage="sendMessage" :joinGroup="joinGroup" :leaveGroup="leaveGroup" />
    </div>
  </div>
</template>

<script>

export default {
  name: 'App',
  data() {
    return {
      socket: null,
      id: "",
      otherConversation: undefined,
      otherConversationMessages: [],
      conversationsList: [],
      userMessages: {}
    }
  },
  methods: {
    onConversationCardClick(user) {
      this.otherConversation = user
      if (!this.userMessages[user.id]) {
        this.userMessages[user.id] = []
      }
      this.otherConversationMessages = this.userMessages[user.id]
    },
    refreshConversationList() {
      this.socket.send(JSON.stringify({ request: "get-conversation-list" }))
    },
    createGroup() {
      this.socket.send(JSON.stringify({ request: "create-group" }))
    },
    joinGroup(groupId) {
      this.socket.send(JSON.stringify({
        request: "join-group",
        data: { senderId: this.id, groupId }
      }))
    },
    leaveGroup(groupId) {
      this.socket.send(JSON.stringify({
        request: "leave-group",
        data: { senderId: this.id, groupId }
      }))
    },
    sendMessage(otherConversation, message) {
      console.log('sendMessage', otherConversation, message)
      const data = {
        senderId: this.id,
        receiverId: otherConversation.id,
        message,
        type: "text"
      }
      this.otherConversationMessages.push(data)
      this.socket.send(JSON.stringify({ request: "text", data }))
    },
    processMessage(msg) {
      const json = JSON.parse(msg)
      console.log("processMessage", json)
      switch (json.type) {
        case "id":
          this.id = json.id
          break
        case "get-conversation-list":
          this.conversationsList = json.conversations
          if (!this.otherConversation) {
            this.otherConversation = this.conversationsList[0]
          }
          break
        case "text":
          const id = json.senderId
          if (!this.userMessages[id]) {
            this.userMessages[id] = []
          }
          this.userMessages[id].push(json)
          break
      }
    },
  },
  mounted() {
    this.socket = new WebSocket(`ws://localhost:8000/connect`)
    this.socket.onopen = () => {
      this.socket.send(JSON.stringify({ request: "get-conversation-list" }))
    }
    this.socket.onmessage = ({ data }) => { this.processMessage(data) }
  },
}
</script>