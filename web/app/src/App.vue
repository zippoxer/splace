<template>
  <div id="app">
    <h3
      class="uk-text-center uk-flex-1">/splace/</h3>
    <hr class="uk-divider-icon">
    <router-view/>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
html,
body {
  background-color: #f8f8f8;
  margin: 0;
  padding: 0;
}
html {
  height: 100%;
}
body {
  min-height: 100%;
  padding: 3em 0;
  box-sizing: border-box;
}
#app {
  width: 800px;
  margin: 0 auto;
  padding-top: 1em;
  padding-bottom: 3em;
  border: solid 1px #ccc;
  background-color: #fff;
  border-radius: 10px;
}
</style>
