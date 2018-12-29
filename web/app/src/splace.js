import axios from 'axios'
import EventEmitter from 'events'

export default class Splace extends EventEmitter {
  constructor () {
    super()
    this.url = window.apiURL || 'http://localhost:30993'
  }

  connect (params) {
    return this._request('POST', '/connect', params)
  }

  search (options) {
    return new EventSource(this.url + '/search?options=' +
      encodeURIComponent(JSON.stringify(options)),
    { retry: null })
  }

  replace (options) {
    return new EventSource(this.url + '/replace?options=' +
      encodeURIComponent(JSON.stringify(options)),
    { retry: null })
  }

  cancel () {
    return this._request('POST', '/cancel')
  }

  _request (method, path, data) {
    let promise = axios({
      method: method,
      data: data,
      url: this.url + path
    })
      .then(resp => resp.data)
    promise.catch(e => this._handleError(e))
    return promise
  }

  _handleError (e) {
    console.error(e)
    this.emit('error', e)
    throw e
  }
}
