<template>
  <div class="uk-container">
    <div class="uk-flex">
      <h4
        v-if="operation.end"
        class="uk-flex-1 uk-margin-remove-bottom">
        Search for "{{ operation.options.search }}" (took {{ took }})
      </h4>
      <h4
        v-else
        class="uk-flex-1 uk-margin-remove-bottom">
        <vk-spinner
          ratio="0.75"
          class="uk-margin-small-right" /> Searching for "{{ operation.options.search }}"
      </h4>
      <div>
        <vk-button
          v-if="operation.end"
          class="uk-button-danger "
          :disabled="selected.length == 0">
          <span v-if="selected.length">
            Replace {{ selected.length }} tables
          </span>
          <span v-else>Replace</span>
        </vk-button>
        <vk-button
          v-else
          class="uk-width-small">Stop</vk-button>
      </div>
    </div>
    <ul class="uk-list uk-list-divider table-list">
      <template
        v-for="(result, resultIndex) in operation.results">
        <li
          v-if="result.rows.length"
          :key="resultIndex"
          :class="{'selected': selected.includes(resultIndex),
                   'expanded': expanded.includes(resultIndex)}"
          @click="expandTable(resultIndex)">
          <div class="uk-flex uk-flex-middle">
            <div class="uk-margin-right">
              <input
                v-model="selected"
                :value="resultIndex"
                @click.stop
                type="checkbox"
                class="uk-checkbox" >
            </div>
            <span class="uk-flex-1">
              {{ result.table }}
            </span>
            <div>
              <i>{{ result.rows.length }} matching rows</i>
              <vk-icon :icon="expanded.includes(resultIndex) ? 'chevron-up' : 'chevron-down'" />
            </div>
          </div>
          <div
            v-if="expanded.includes(resultIndex)"
            class="uk-overflow-auto table-rows">
            <table
              class="uk-table uk-table-small uk-table-divider"
              @click.stop>
              <thead>
                <tr>
                  <th
                    v-for="(column, i) in tables[result.table]"
                    :key="i">{{ column }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(row, i) in result.rows"
                  :key="i">
                  <td
                    v-for="(value, j) in row"
                    :key="j">
                    <span
                      v-for="(v, i) in highlightCache[resultIndex][i][j]"
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
        let rows = this.operation.results[index].rows
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
      return Number(ms / 1e3).toFixed(1) + 's'
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
    &.expanded {
      background-color: #f8f8f8;
    }
    &.selected {
      background-color: #ffffdd;
    }
  }

  .table-rows {
    margin-top: 10px;
    background-color: #fff;
    max-height: 350px;
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
