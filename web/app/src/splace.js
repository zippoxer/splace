import axios from 'axios'

export default class Splace {
  connect (params) {
    return this.request('POST', '/connect', params)
  }

  request (method, path, data) {
    return axios({
      method: method,
      data: data,
      url: path
    })
  }
}
