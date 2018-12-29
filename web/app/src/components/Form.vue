<template>
  <div>
    <!-- Search/Replace inputs -->
    <div class="uk-container">
      <form>
        <vk-grid>
          <div class="uk-width-1-2@s">
            <div class="uk-inline uk-width-1-1">
              <vk-icon
                class="uk-form-icon"
                icon="search"/>
              <input
                v-model="options.search"
                class="uk-input"
                type="text"
                placeholder="Search...">
              <div
                class="uk-position-center-right uk-margin-small-right"
                style="font-family: Consolas; cursor: pointer; user-select: none"
                @click="toggleSearchMode">
                {{ consts.SEARCH_MODES[options.mode] }}
              </div>
            </div>
          </div>
          <div class="uk-width-1-2@s">
            <div class="uk-inline uk-width-1-1">
              <vk-icon
                class="uk-form-icon"
                icon="pencil" />
              <input
                v-model="options.replace"
                class="uk-input"
                type="text"
                placeholder="Replace...">
            </div>
          </div>
        </vk-grid>
      </form>
    </div>

    <hr class="uk-divider-icon">
    <DatabaseForm v-model="options.db" />
    <hr class="uk-divider-icon">

    <!-- Search/Replace buttons -->
    <div class="uk-container">
      <vk-grid>
        <div class="uk-width-1-2@s">
          <vk-button
            class="uk-button-primary uk-width-1"
            @click="search">Search</vk-button>
        </div>
        <div class="uk-width-1-2@s">
          <vk-button class="uk-button-danger uk-width-1">Search & Replace</vk-button>
        </div>
      </vk-grid>
    </div>

    <template v-if="operation">
      <hr class="uk-divider-icon">
      <SearchResults
        v-if="operation.kind == 'search'"
        :operation="operation"
        :tables="tables" />
        <!-- <ReplaceResults v-if="operation.kind == 'replace'" /> -->
    </template>
  </div>
</template>

<script>
import * as consts from '../consts'
import DatabaseForm from './DatabaseForm'
import SearchResults from './SearchResults'

export default {
  name: 'Form',
  components: {DatabaseForm, SearchResults},
  data () {
    return {
      consts,
      operation: null,
      options: {
        search: 'quizard',
        replace: '',
        mode: 0,

        db: {
          host: '',
          database: '',
          user: '',
          password: '',
          engine: 0,
          driver: 'direct'
        }
      }
    }
  },
  methods: {
    toggleSearchMode () {
      let mode = this.options.mode + 1
      if (mode >= consts.SEARCH_MODES.length) {
        mode = 0
      }
      this.options.mode = mode
    },
    search () {
      this.operation = null
      this.connect().then(() => {
        let options = JSON.parse(JSON.stringify(this.options)) // :-(
        this.operation = {
          kind: 'search',
          start: new Date(),
          end: null,
          options,
          results: []
        }

        let {mode, search} = options
        if (mode === 0) {
          mode = 1
          search = '%' + search.replace('%', '\\%') + '%'
        }
        var searcher = this.$splace.search({
          Search: search,
          Mode: mode,
          Tables: this.tables,
          Limit: 1000
        })
        searcher.addEventListener('table', e => {
          let data = JSON.parse(e.data)
          this.operation.results.push({
            table: data.Table,
            sql: data.SQL,
            columns: data.Columns,
            rows: [],
            start: data.Start
          })
        })
        searcher.addEventListener('rows', e => {
          let rows = JSON.parse(e.data)
          let index = this.operation.results.length - 1
          this.operation.results[index].rows = this.operation.results[index].rows.concat(rows)
        })
        searcher.addEventListener('done', e => {
          searcher.close()
          this.operation.end = new Date()
        })
      })
    },
    connect () {
      let db = this.options.db
      return this.$splace.connect({
        Driver: db.driver,
        Engine: db.engine,
        Host: db.host,
        Database: db.database,
        User: db.user
      }).then(resp => {
        this.tables = resp.Tables
        return resp
      })
    }
  }
}
</script>

<style scoped>
</style>
