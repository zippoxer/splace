<template>
  <div>
    <vk-button @click="connect">asd</vk-button>
    <dl class="uk-description-list">
        <template v-for="(columns, table) in tables">
          <dt :key="table">{{table}}</dt>
          <dd :key="table">{{columns.join(", ")}}</dd>
        </template>
    </dl>
  </div>
</template>

<script>
export default {
  name: 'HelloWorld',
  created () {
  },
  data () {
    return {
      tables: null
    }
  },
  methods: {
    connect () {
      console.log('connecting!')
      this.$splace.connect({
        Driver: 'direct',
        Engine: 0,
        Host: 'localhost',
        Database: 'quizard_web_dev',
        User: 'root'
      }).then(resp => {
        this.tables = resp.data.Tables
      })
    }
  }
}
</script>

<style scoped>
h1,
h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
