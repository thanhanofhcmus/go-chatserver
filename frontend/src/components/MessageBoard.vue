<script setup>
import UserMessage from './UserMessage.vue';
import OtherMessage from './OtherMessage.vue';
import SendIcon from './icons/Send.vue';
</script>
<template>
  <div class="col-span-2 block" v-if="!otherConversation"></div>
  <div class="col-span-2 block" v-else>
    <div class="w-full">
      <div class="relative flex items-center p-3 border-b border-gray-300">
        <span class="block ml-2 font-bold text-gray-600">{{ otherConversation.id }}</span>
        <template v-if="otherConversation.type === 'group'">
          <button class="border bg-green-600 text-gray-200 font-bold m-2 p-2 rounded-lg"
            @click="() => joinGroup(otherConversation.id)">Join Group
          </button>
          <button class="border bg-red-600 text-gray-200 font-bold m-2 p-2 rounded-lg"
            @click="() => leaveGroup(otherConversation.id)">Leave Group
          </button>
        </template>
      </div>
      <div class="relative w-full p-6 overflow-y-auto h-[40rem]">
        <li class="space-y-2">
          <template v-for="msg in messages">
            <UserMessage v-if="msg.senderId === id" :text="msg.message" />
            <OtherMessage v-else :text="msg.message" :id="msg.senderId"/>
          </template>
        </li>
      </div>

      <div class="flex items-center justify-between w-full p-3 border-t border-gray-300">
        <input type="text" placeholder="Message"
          class="block w-full py-2 pl-4 mx-3 bg-gray-100 rounded-full outline-none focus:text-gray-700" name="message"
          v-model="text" @keyup.enter="onSendClick" required />
        <button type="submit" @click="onSendClick">
          <SendIcon />
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "MessageBoard",
  props: ["id", "otherConversation", "messages", "sendMessage", "joinGroup", "leaveGroup"],
  data() {
    return { text: "" }
  },
  methods: {
    onSendClick() {
      this.sendMessage(this.otherConversation, this.text)
      this.text = ""
    }
  }
}
</script>