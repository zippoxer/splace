<template>
  <div>
    <form @submit="search">
      <!-- Search/Replace inputs -->
      <div class="uk-container">
        <vk-grid>
          <div class="uk-width-1-2@s">
            <div class="uk-inline uk-width-1-1">
              <vk-icon class="uk-form-icon" icon="search" />
              <input
                v-model="options.search"
                ref="search"
                class="uk-input"
                type="text"
                placeholder="Search..."
                autofocus
              />
              <div
                class="uk-position-center-right uk-margin-small-right"
                style="font-family: Consolas; cursor: pointer; user-select: none"
                @click="toggleSearchMode"
              >{{ consts.SEARCH_MODES[options.mode] }}</div>
            </div>
          </div>
          <div class="uk-width-1-2@s">
            <div class="uk-inline uk-width-1-1">
              <vk-icon class="uk-form-icon" icon="pencil" />
              <input
                v-model="options.replace"
                ref="replace"
                class="uk-input"
                type="text"
                placeholder="Replace..."
              />
            </div>
          </div>
        </vk-grid>
      </div>

      <hr class="uk-divider-icon" />
      <DatabaseForm v-model="options.db" @check="checkPhpProxy" :status="dbStatus" />
      <hr class="uk-divider-icon" />

      <!-- Search/Replace buttons -->
      <div class="uk-container">
        <vk-grid>
          <div class="uk-width-1-2@s">
            <vk-button html-type="submit" class="uk-button-primary uk-width-1">Search</vk-button>
          </div>
          <div class="uk-width-1-2@s">
            <vk-button class="uk-button-danger uk-width-1" @click="replace">Search & Replace</vk-button>
          </div>
        </vk-grid>
      </div>
    </form>

    <div v-if="alerts.length" class="uk-container uk-margin-top">
      <div v-for="(alert, i) in alerts" :key="i" :class="`uk-alert-danger uk-alert`" uk-alert>
        <a @click="dismissAlert(i)" class="uk-alert-close uk-close uk-icon" uk-close>
          <svg
            width="14"
            height="14"
            viewBox="0 0 14 14"
            xmlns="http://www.w3.org/2000/svg"
            data-svg="close-icon"
          >
            <line fill="none" stroke="#000" stroke-width="1.1" x1="1" y1="1" x2="13" y2="13" />
            <line fill="none" stroke="#000" stroke-width="1.1" x1="13" y1="1" x2="1" y2="13" />
          </svg>
        </a>
        <p>
          <vk-icon icon="warning" />
          {{ alert.message }}
        </p>
      </div>
    </div>

    <template v-if="discoveredConfigs.length">
      <hr class="uk-divider-icon" />
      <div class="uk-container">
        <div class="uk-width-1 uk-flex uk-flex-wrap">
          <vk-card
            class="uk-width-1-1@s"
            padding="small"
            v-for="(config, i) in discoveredConfigs"
            :key="i"
          >
            <vk-label slot="badge">{{ config.Who }}</vk-label>
            <div slot="header">
              <vk-card-title
                class="uk-margin-remove-bottom"
              >Discovered `{{ config.Config.Database }}`</vk-card-title>
              <p
                class="uk-text-meta uk-margin-remove-top uk-margin-remove-bottom"
              >{{ config.Where }}</p>
            </div>
            <div slot="footer">
              <vk-button-link type="text" @click="useDiscoveredConfig(i)">Connect</vk-button-link>
            </div>
          </vk-card>
        </div>
      </div>
    </template>

    <template v-if="currentSearch || currentReplace">
      <hr class="uk-divider-icon" />
      <vk-tabs-vertical align="left" class="result-tabs">
        <vk-tabs-item title icon="search" :disabled="!currentSearch">
          <SearchResult v-if="currentSearch" :operation="currentSearch" :tables="tables" />
        </vk-tabs-item>
        <vk-tabs-item title icon="pencil" :disabled="!currentReplace">
          <ReplaceResult v-if="currentReplace" :operation="currentReplace" :tables="tables" />
        </vk-tabs-item>
      </vk-tabs-vertical>
    </template>
  </div>
</template>

<script>
import * as consts from "../consts";
import DatabaseForm from "./DatabaseForm";
import SearchResult from "./SearchResult";
import ReplaceResult from "./ReplaceResult";

export default {
  name: "Form",
  components: { DatabaseForm, SearchResult, ReplaceResult },
  data() {
    return {
      consts,
      alerts: [],

      dbStatus: "",
      tables: {},
      discoveredConfigs: [],

      currentSearch: null,
      currentReplace: null,

      options: {
        search: "",
        replace: "",
        mode: Object.keys(consts.SEARCH_MODES)[0],

        db: {
          host: "",
          database: "",
          user: "",
          password: "",
          engine: "mysql",
          driver: "direct"
          // driver: 'php',
          // url: 'http://localhost/splace-proxy.php'
        }
      }
    };
  },
  mounted() {
    document.onkeypress = ev => {
      // Focus search on slash press.
      switch (ev.key) {
        case "/":
        case "\\":
          let tag = document.activeElement.tagName.toLowerCase();
          if (tag !== "textarea" && tag !== "input") {
            ev.preventDefault();
            if (ev.key === "/") {
              this.$refs.search.focus();
            } else {
              this.$refs.replace.focus();
            }
          }
          break;
      }
    };
  },
  methods: {
    toggleSearchMode() {
      let mode = Number(this.options.mode) + 1;
      if (!consts.SEARCH_MODES[mode]) {
        mode = Object.keys(consts.SEARCH_MODES)[0];
      }
      this.options.mode = mode;
    },
    search() {
      this.currentSearch = null;
      this.currentReplace = null;

      let options = JSON.parse(JSON.stringify(this.options)); // :-(
      this.$nextTick(() => {
        this.currentSearch = {
          kind: "search",
          start: new Date(),
          end: null,
          options,
          cancel: () => {},
          result: {
            tables: {},
            rows: {},
            totalRows: 0
          }
        };

        this.$nextTick(() => {
          document
            .querySelector(".result-tabs ul.uk-tab li:first-child a")
            .click();
        });

        this.connect()
          .then(() => {
            let lastUpdate = null;
            let searcher = this.$splace.search({
              Search: options.search,
              Mode: Number(options.mode),
              Tables: this.tables,
              Limit: 0
            });
            this.currentSearch.cancel = searcher.cancel;
            searcher.addEventListener("table", e => {
              let data = JSON.parse(e.data);
              this.currentSearch.result.rows[data.Table] = [];
              let table = {
                table: data.Table,
                sql: data.SQL,
                columns: data.Columns,
                totalRows: 0,
                start: data.Start
              };
              if (lastUpdate === null || new Date() - lastUpdate > 100) {
                this.$set(this.currentSearch.result.tables, data.Table, table);
                lastUpdate = new Date();
              } else {
                this.currentSearch.result.tables[data.Table] = table;
              }
            });
            searcher.addEventListener("rows", e => {
              let data = JSON.parse(e.data);
              let table = data[0];
              let rowCount = data[1];
              this.currentSearch.result.tables[table].totalRows += rowCount;
              this.currentSearch.result.totalRows += rowCount;
              if (data.length === 3) {
                let rows = data[2];
                let newRows = this.currentSearch.result.rows[table].concat(
                  rows
                );
                this.currentSearch.result.rows[table] = newRows;
              }
            });
            searcher.addEventListener("done", e => {
              searcher.close();

              this.$set(this.currentSearch, "end", new Date());

              var data = JSON.parse(e.data);
              if (data.Error) {
                this.pushAlert(data.Error);
              }
            });
            searcher.addEventListener("cancel", e => {
              this.$set(this.currentSearch, "end", new Date());
            });
            searcher.onerror = e => {
              searcher.close();
              console.error(e);
            };
          })
          .catch(e => {
            this.currentSearch = null;
            this.currentReplace = null;
          });
      });
    },
    replace() {
      this.currentReplace = null;

      this.$nextTick(() => {
        let options = JSON.parse(JSON.stringify(this.options)); // :-(
        this.currentReplace = {
          kind: "replace",
          start: new Date(),
          end: null,
          options,
          result: {
            tables: {},
            totalAffectedRows: 0
          }
        };

        this.$nextTick(() => {
          document
            .querySelector(".result-tabs ul.uk-tab li:last-child a")
            .click();
        });

        this.connect()
          .catch(e => {
            this.currentReplace = null;
          })
          .then(() => {
            var replacer = this.$splace.replace({
              Search: options.search,
              Replace: options.replace,
              Mode: Number(options.mode),
              Tables: this.tables,
              Limit: 0
            });
            replacer.addEventListener("table", e => {
              let data = JSON.parse(e.data);
              this.currentReplace.result.tables[data.Table] = {
                table: data.Table,
                sql: data.SQL,
                columns: data.Columns,
                affectedRows: 0,
                start: data.Start
              };
            });
            replacer.addEventListener("affected_rows", e => {
              let data = JSON.parse(e.data);
              let table = data[0];
              let affectedRows = data[1];
              this.currentReplace.result.tables[
                table
              ].affectedRows += affectedRows;
              this.currentReplace.result.totalAffectedRows += affectedRows;
            });
            replacer.addEventListener("done", e => {
              replacer.close();

              this.$set(this.currentReplace, "end", new Date());

              var data = JSON.parse(e.data);
              if (data.Error) {
                this.pushAlert(data.Error);
              }
            });
            replacer.onerror = e => {
              replacer.close();
              console.error(e);
            };
          });
      });
    },
    connect() {
      this.resetAlerts();
      this.dbStatus = "connecting";

      let db = this.options.db;
      return this.$splace
        .connect({
          Driver: db.driver,
          Engine: db.engine,
          Host: db.host,
          Database: db.database,
          User: db.user,
          Pwd: db.password,
          URL: db.url
        })
        .then(resp => {
          this.tables = resp.Tables || {};

          for (let i in resp.DiscoveredConfigs) {
            let cfg = resp.DiscoveredConfigs[i].Config;
            resp.DiscoveredConfigs[i].dbOptions = {
              ...this.options.db,
              host: cfg.Host,
              database: cfg.Database,
              user: cfg.User,
              password: cfg.Pwd,
              engine: cfg.Engine,
              driver: cfg.Driver,
              url: cfg.URL
            };
          }

          if (Array.isArray(resp.DiscoveredConfigs)) {
            this.discoveredConfigs = resp.DiscoveredConfigs.filter(c => {
              return (
                JSON.stringify(c.dbOptions) !== JSON.stringify(this.options.db)
              );
            });
          }

          if (resp.Error) {
            let e = new Error(resp.Error);
            e.response = { data: resp };
            throw e;
          }

          this.dbStatus = "connected";
          return resp;
        })
        .catch(e => {
          console.error(e);
          this.dbStatus = "error";
          this.pushAlert(e.response.data.Error);
          throw e;
        });
    },
    resetAlerts() {
      this.alerts = [];
    },
    dismissAlert(index) {
      this.alerts.splice(index, 1);
    },
    pushAlert(message) {
      this.alerts.push({
        message
      });
    },
    checkPhpProxy() {
      this.connect();
    },
    useDiscoveredConfig(i) {
      let cfg = this.discoveredConfigs[i];
      this.discoveredConfigs.splice(i, 1);
      this.options.db = cfg.dbOptions;
    }
  }
};
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

.uk-notification {
  width: 500px;
  &.uk-notification-bottom-center {
    margin-left: -250px;
  }
  .uk-notification-message {
    background-color: #fff;
    -webkit-box-shadow: 0px 2px 20px -8px rgba(0, 0, 0, 0.35);
    -moz-box-shadow: 0px 2px 20px -8px rgba(0, 0, 0, 0.35);
    box-shadow: 0px 2px 20px -8px rgba(0, 0, 0, 0.35);
    font-size: 18px;
    padding: 15px 30px;
  }
}
</style>
