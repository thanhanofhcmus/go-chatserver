<template>
  <li class="flex justify-start">
    <div class="flex flex-col">
      <span v-if="id">{{ id }}</span>
      <div class="relative max-w-xl px-4 py-2 text-gray-700 rounded shadow">
        <span class="block">{{ text }}</span>
      </div>
    </div>
  </li>
</template>

<script>
export default {
  props: ["text", "id"]
}
</script>