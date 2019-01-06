<template>
  <div>
    <form @submit="search">
      <!-- Search/Replace inputs -->
      <div class="uk-container">
        <vk-grid>
          <div class="uk-width-1-2@s">
            <div class="uk-inline uk-width-1-1">
              <vk-icon
                class="uk-form-icon"
                icon="search"/>
              <input
                v-model="options.search"
                ref="search"
                class="uk-input"
                type="text"
                placeholder="Search..."
                autofocus>
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
                ref="replace"
                class="uk-input"
                type="text"
                placeholder="Replace...">
            </div>
          </div>
        </vk-grid>
      </div>

      <hr class="uk-divider-icon">
      <DatabaseForm v-model="options.db" />
      <hr class="uk-divider-icon">

      <!-- Search/Replace buttons -->
      <div class="uk-container">
        <vk-grid>
          <div class="uk-width-1-2@s">
            <vk-button
              html-type="submit"
              class="uk-button-primary uk-width-1">Search</vk-button>
          </div>
          <div class="uk-width-1-2@s">
            <vk-button
              class="uk-button-danger uk-width-1"
              @click="replace">Search & Replace</vk-button>
          </div>
        </vk-grid>
      </div>
    </form>

    <template v-if="currentSearch || currentReplace">
      <hr class="uk-divider-icon">
      <vk-tabs-vertical
        align="left"
        class="result-tabs">
        <vk-tabs-item
          title=""
          icon="search">
          <SearchResult
            v-if="currentSearch"
            :operation="currentSearch"
            :tables="tables" />
        </vk-tabs-item>
        <vk-tabs-item
          title=""
          icon="pencil"
          :disabled="!currentReplace">
          <ReplaceResult
            v-if="currentReplace"
            :operation="currentReplace"
            :tables="tables" />
        </vk-tabs-item>
      </vk-tabs-vertical>
    </template>
  </div>
</template>

<script>
import * as consts from '../consts'
import DatabaseForm from './DatabaseForm'
import SearchResult from './SearchResult'
import ReplaceResult from './ReplaceResult'

export default {
  name: 'Form',
  components: {DatabaseForm, SearchResult, ReplaceResult},
  data () {
    return {
      consts,
      tables: {},
      currentSearch: null,
      currentReplace: null,
      options: {
        search: 'quizard',
        replace: 'qquizzard',
        mode: Object.keys(consts.SEARCH_MODES)[0],

        db: {
          host: 'localhost',
          database: 'quizard_web_dev',
          user: 'root',
          password: '',
          engine: 0,
          // driver: 'direct',
          driver: 'php',
          url: 'http://localhost/splace-proxy.php'
        }
      }
    }
  },
  mounted () {
    document.onkeypress = ev => {
      // Focus search on slash press.
      switch (ev.key) {
        case '/':
        case '\\':
          let tag = document.activeElement.tagName.toLowerCase()
          if (tag !== 'textarea' && tag !== 'input') {
            ev.preventDefault()
            if (ev.key === '/') {
              this.$refs.search.focus()
            } else {
              this.$refs.replace.focus()
            }
          }
          break
      }
    }
  },
  methods: {
    toggleSearchMode () {
      let mode = Number(this.options.mode) + 1
      if (!consts.SEARCH_MODES[mode]) {
        mode = Object.keys(consts.SEARCH_MODES)[0]
      }
      this.options.mode = mode
    },
    search () {
      this.currentSearch = null
      this.currentReplace = null
      let options = JSON.parse(JSON.stringify(this.options)) // :-(
      this.currentSearch = {
        kind: 'search',
        start: new Date(),
        end: null,
        options,
        cancel: () => {},
        result: {
          tables: {},
          rows: {},
          totalRows: 0
        }
      }

      this.$nextTick(() => {
        document.querySelector('.result-tabs ul.uk-tab li:first-child a').click()
      })

      this.connect().then(() => {
        let lastUpdate = null
        let searcher = this.$splace.search({
          Search: options.search,
          Mode: Number(options.mode),
          Tables: this.tables,
          Limit: 0
        })
        this.currentSearch.cancel = searcher.cancel
        searcher.addEventListener('table', e => {
          let data = JSON.parse(e.data)
          this.currentSearch.result.rows[data.Table] = []
          let table = {
            table: data.Table,
            sql: data.SQL,
            columns: data.Columns,
            totalRows: 0,
            start: data.Start
          }
          if (lastUpdate === null || new Date() - lastUpdate > 100) {
            this.$set(this.currentSearch.result.tables, data.Table, table)
            lastUpdate = new Date()
          } else {
            this.currentSearch.result.tables[data.Table] = table
          }
        })
        searcher.addEventListener('rows', e => {
          let data = JSON.parse(e.data)
          let table = data[0]
          let rowCount = data[1]
          this.currentSearch.result.tables[table].totalRows += rowCount
          this.currentSearch.result.totalRows += rowCount
          if (data.length === 3) {
            let rows = data[2]
            let newRows = this.currentSearch.result.rows[table].concat(rows)
            this.currentSearch.result.rows[table] = newRows
          }
        })
        searcher.addEventListener('done', e => {
          searcher.close()
          this.$set(this.currentSearch, 'end', new Date())
        })
        searcher.addEventListener('cancel', e => {
          this.$set(this.currentSearch, 'end', new Date())
        })
        searcher.onerror = (e) => {
          searcher.close()
          console.error(e)
        }
      })
    },
    replace () {
      this.currentReplace = null
      let options = JSON.parse(JSON.stringify(this.options)) // :-(
      this.currentReplace = {
        kind: 'replace',
        start: new Date(),
        end: null,
        options,
        result: {
          tables: {},
          totalAffectedRows: 0
        }
      }

      this.$nextTick(() => {
        document.querySelector('.result-tabs ul.uk-tab li:last-child a').click()
      })

      this.connect().then(() => {
        var replacer = this.$splace.replace({
          Search: options.search,
          Replace: options.replace,
          Mode: Number(options.mode),
          Tables: this.tables,
          Limit: 0
        })
        replacer.addEventListener('table', e => {
          let data = JSON.parse(e.data)
          this.currentReplace.result.tables[data.Table] = {
            table: data.Table,
            sql: data.SQL,
            columns: data.Columns,
            affectedRows: 0,
            start: data.Start
          }
        })
        replacer.addEventListener('affected_rows', e => {
          let data = JSON.parse(e.data)
          let table = data[0]
          let affectedRows = data[1]
          this.currentReplace.result.tables[table].affectedRows += affectedRows
          this.currentReplace.result.totalAffectedRows += affectedRows
        })
        replacer.addEventListener('done', e => {
          replacer.close()
          this.currentReplace.end = new Date()
        })
        replacer.onerror = (e) => {
          replacer.close()
          console.error(e)
        }
      })
    },
    connect () {
      let db = this.options.db
      return this.$splace.connect({
        Driver: db.driver,
        Engine: db.engine,
        Host: db.host,
        Database: db.database,
        User: db.user,
        Pwd: db.password,
        URL: db.url
      }).then(resp => {
        this.tables = resp.Tables
        return resp
      })
    }
  }
}
</script>

<style lang="scss">
.result-tabs {
  .uk-tab > li:last-child.uk-active > a {
    border-color: #f0506e;
  }
  > .uk-width-expand {
    padding-left: 0 !important;
  }
}
</style>
