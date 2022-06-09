<template>
  <li class="flex justify-end">
      <div class="relative max-w-xl px-4 py-2 text-gray-700 bg-gray-100 rounded shadow">
        <span class="block">{{ text }}</span>
      </div>
  </li>
</template>

<script>
export default {
  props: ["text", "id"]
}
</script>