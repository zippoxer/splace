<template>
  <div class="uk-container">
    <div class="uk-flex">
      <h4
        v-if="operation.end"
        class="uk-flex-1 uk-margin-remove-bottom">
        Search for "{{ operation.options.search }}" ({{ took }} seconds)
      </h4>
      <h4
        v-else
        class="uk-flex-1 uk-margin-remove-bottom">
        Searching for "{{ operation.options.search }}"
      </h4>
      <div v-if="!noResults">
        <vk-button
          v-if="operation.end"
          icon="search"
          class="uk-button-danger "
          :disabled="selected.length == 0">
          <span v-if="selected.length">
            Replace {{ selected.length }} tables
          </span>
          <span v-else>Replace</span>
        </vk-button>
        <vk-button
          v-else
          @click="operation.cancel"
          class="uk-width-small uk-inline">
          <vk-spinner
            ratio="0.55"
            class="uk-position-center-left uk-position-small" />
          Stop
        </vk-button>
      </div>
    </div>
    <p
      v-if="noResults"
      class="uk-text-muted">😔 No results found.</p>
    <ul
      v-else-if="operation.result.totalRows"
      class="uk-list uk-list-divider table-list">
      <template
        v-for="(table, tableName) in operation.result.tables">
        <li
          v-if="operation.result.rows[tableName].length"
          :key="tableName"
          :class="{'selected': selected.includes(tableName),
                   'expanded': expanded.includes(tableName)}"
          @click="expandTable(tableName)">
          <div class="uk-flex uk-flex-middle">
            <label
              class="replace-checkbox"
              @click.stop>
              <input
                v-model="selected"
                :value="tableName"
                type="checkbox"
                class="uk-checkbox" >
            </label>
            <span class="uk-flex-1 table-name">
              {{ tableName }}
            </span>
            <div>
              <i>{{ table.totalRows }} matching rows</i>
              <vk-icon :icon="expanded.includes(tableName) ? 'chevron-up' : 'chevron-down'" />
            </div>
          </div>
          <div
            v-if="expanded.includes(tableName)"
            class="uk-overflow-auto table-rows">
            <table
              class="uk-table uk-table-small uk-table-divider"
              @click.stop>
              <thead>
                <tr>
                  <th
                    v-for="(column, i) in tables[tableName]"
                    :key="i">{{ column.Column }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(row, i) in operation.result.rows[tableName]"
                  :key="i">
                  <td
                    v-for="(value, j) in row"
                    :key="j">
                    <span
                      v-for="(v, i) in highlightCache[tableName][i][j]"
                      :key="i"
                      :class="{'hl': v[0]}">{{ v[1] }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </li>
      </template>
    </ul>
  </div>
</template>

<script>
export default {
  name: 'SearchResults',
  props: {
    operation: {
      type: Object,
      required: true
    },
    tables: {
      type: Object,
      required: true
    }
  },
  data: function () {
    return {
      selected: [],
      expanded: [],
      highlightCache: {}
    }
  },
  methods: {
    expandTable (index) {
      let i = this.expanded.indexOf(index)
      if (i !== -1) {
        this.expanded.splice(i, 1)
      } else {
        this.expanded.push(index)
        let table = this.operation.result.tables[index].table
        let rows = this.operation.result.rows[table]
        this.highlightCache[index] = rows.map(row => row.map(v => this.highlightRow(v)))
      }
    },
    highlightRow (value) {
      let fragments = []
      while (true) {
        let i = value.indexOf(this.operation.options.search)
        if (i === -1) {
          break
        }
        if (i > 0) {
          fragments.push([false, value.slice(0, i)])
        }
        fragments.push([true, value.slice(i, i + this.operation.options.search.length)])
        value = value.slice(i + this.operation.options.search.length)
      }
      if (value.length > 0) {
        fragments.push([false, value])
      }
      return fragments
    }
  },
  computed: {
    took () {
      let ms = this.operation.end - this.operation.start
      return Number(ms / 1e3).toFixed(2)
    },
    noResults () {
      return this.operation.end && this.operation.result.totalRows === 0
    }
  }
}
</script>

<style lang="scss" scoped>
.table-list {
  border-top: 1px solid #e5e5e5;
  border-bottom: 1px solid #e5e5e5;

  li {
    margin: 0 !important;
    padding: 10px !important;
    cursor: pointer;
    position: relative;
    &.expanded {
      background-color: #f8f8f8;
    }
    &.selected {
      background-color: #ffffdd;
    }
    .replace-checkbox {
      position: absolute;
      left: 0;
      top: 0;
      width: 40px;
      height: 43.5px;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;

      input {
        margin-top: 2px;
      }
    }
    .table-name {
      margin-left: 30px;
    }
  }

  .table-rows {
    background-color: #fff;
    max-height: 350px;
    margin-top: 10px;
    cursor: default;

    table {
      td {
        font-size: 13px;
        white-space: nowrap;

        span.hl {
          background-color: #ffffdd;
        }
      }
    }
  }
}
</style>
