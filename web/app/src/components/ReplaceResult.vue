<template>
  <div class="uk-container">
    <div class="uk-flex">
      <h4
        v-if="operation.end"
        class="uk-flex-1 uk-margin-remove-bottom">
        Replaced "{{ operation.options.search }}" with "{{ operation.options.replace }}" ({{ took }} seconds)
      </h4>
      <h4
        v-else
        class="uk-flex-1 uk-margin-remove-bottom">
        Replacing "{{ operation.options.search }}" with "{{ operation.options.replace }}"
      </h4>
      <div v-if="!operation.end">
        <vk-button
          class="uk-width-small uk-inline">
          <vk-spinner
            ratio="0.55"
            class="uk-position-center-left uk-position-small" />
          Stop
        </vk-button>
      </div>
    </div>
    <p
      v-if="operation.end && operation.result.totalAffectedRows === 0"
      class="uk-text-muted">No replacements were made.</p>
    <ul
      v-else-if="operation.result.totalAffectedRows"
      class="uk-list uk-list-divider table-list">
      <template
        v-for="(table, tableName) in operation.result.tables">
        <li
          v-if="table.affectedRows"
          :key="tableName"
          @click="expandTable(tableIndex)">
          <div class="uk-flex uk-flex-middle">
            <vk-icon
              icon="table"
              class="uk-margin-small-right" />
            <span class="uk-flex-1 table-name">
              {{ tableName }}
            </span>
            <div>
              <i>{{ table.affectedRows }} rows replaced</i>
            </div>
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
  computed: {
    took () {
      let ms = this.operation.end - this.operation.start
      return Number(ms / 1e3).toFixed(2)
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
    position: relative;
  }
}
</style>
