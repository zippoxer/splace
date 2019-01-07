<template>
  <div class="uk-container">
    <form class="uk-form-stacked uk-margin">
      <div class="uk-flex uk-middle">
        <div class="uk-flex-1 uk-flex uk-flex-middle">
          <legend class="uk-text-lead uk-width-auto">MySQL</legend>
          <vk-icon icon="triangle-down"/>
        </div>
        <vk-button-link
          type="link"
          href="/dump"
          tabindex="-1">Download backup</vk-button-link>
      </div>
      <vk-grid
        class="uk-margin"
        margin="uk-grid-margin-small">
        <div class="uk-width-1-2@s">
          <div class="uk-form-controls">
            <div class="uk-inline uk-width-1-1">
              <vk-icon
                class="uk-form-icon"
                icon="server" />
              <input
                v-model="value.host"
                class="uk-input"
                placeholder="Host"
                type="text">
            </div>
          </div>
        </div>
        <div class="uk-width-1-2@s">
          <div class="uk-form-controls">
            <div class="uk-inline uk-width-1-1">
              <vk-icon
                class="uk-form-icon"
                icon="database" />
              <input
                v-model="value.database"
                class="uk-input"
                placeholder="Database Name"
                type="text">
            </div>
          </div>
        </div>
        <div class="uk-width-1-2@s">
          <div class="uk-form-controls">
            <div class="uk-inline uk-width-1-1">
              <vk-icon
                class="uk-form-icon"
                icon="user" />
              <input
                v-model="value.user"
                class="uk-input"
                placeholder="User"
                type="text">
            </div>
          </div>
        </div>
        <div class="uk-width-1-2@s">
          <div class="uk-form-controls">
            <div class="uk-inline uk-width-1-1">
              <vk-icon
                class="uk-form-icon"
                icon="lock" />
              <input
                v-model="value.password"
                class="uk-input"
                placeholder="Password"
                type="password">
            </div>
          </div>
        </div>
      </vk-grid>
    </form>
    <div class="uk-flex">
      <form class="uk-form-stacked uk-margin-medium-right">
        <label class="uk-form-label">Connection</label>
        <select
          class="uk-select uk-width-auto"
          v-model="value.driver">
          <option
            v-for="(label, key) in consts.DB_DRIVERS"
            :key="key"
            :value="key">{{ label }}</option>
        </select>
      </form>
      <template v-if="value.driver == 'php'">
        <form
          class="uk-form-stacked uk-margin-medium-right uk-flex-none">
          <div>
            <label
              for=""
              class="uk-form-label">
              1. Download proxy script
            </label>
            <div class="uk-form-controls">
              <div class="uk-inline">
                <vk-icon
                  class="uk-form-icon"
                  icon="download" />
                <vk-button-link href="/download-php-proxy">
                  Download
                </vk-button-link>
              </div>
            </div>
          </div>
        </form>
        <form
          class="uk-form-stacked uk-flex-1">
          <div>
            <label
              for=""
              class="uk-form-label">
              2. Drag it to your website's public folder and fill it's URL:
            </label>
            <div class="uk-form-controls">
              <div class="uk-inline uk-width-1">
                <vk-icon
                  class="uk-form-icon"
                  icon="world"/>
                <input
                  v-model="value.url"
                  type="text"
                  class="uk-input uk-width-1"
                  placeholder="http://example.com/splace-proxy.php" >
              </div>
            </div>
          </div>
        </form>
      </template>
    </div>
  </div>
</template>

<script>
import * as consts from '../consts'

export default {
  name: 'DatabaseForm',
  props: {
    value: {
      type: Object,
      required: true
    }
  },
  data: function () {
    return {
      consts
    }
  }
}
</script>
