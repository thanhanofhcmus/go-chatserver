<template>
  <div class="border-r border-gray-300 col-span-1">
    <button class="border bg-red-600 text-gray-200 font-bold m-2 p-2 rounded-lg"
    @click="refreshConversationList">refresh
    </button>
    <button class="border bg-green-600 text-gray-200 font-bold m-2 p-2 rounded-lg"
    @click="createGroup">create group
    </button>

    <div class="flex justify-center items-center font-bold text-sm text-gray-800 my-2">{{ id }}</div>
    <ul class="overflow-auto h-[32rem]">
      <h2 class="my-2 mb-2 ml-2 text-lg text-gray-600">Chats</h2>
      <li v-for="conv in conversationsList">
        <div
          v-if="id !== conv.id"
          class="border flex items-center px-3 py-2 text-sm bg-gray-100 border-b border-gray-300 hover:bg-gray-400"
          @click="() => onConversationCardClick(conv)">
          <div class="w-full pb-2">
            <div class="flex justify-between">
              <div v-if="conv.type === 'group'" >GR </div>
              <span class="block ml-2 font-semibold text-gray-600">{{ conv.id }}</span>
            </div>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  props: ["id", "conversationsList", "onConversationCardClick", "refreshConversationList", "createGroup"],
}
</script>