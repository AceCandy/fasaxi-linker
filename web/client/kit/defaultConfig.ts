const DEFAULT_CONFIG = 'hlink_default_config'

const DEFAULT_TEMPLATE = `export default {
  include: [],
  exclude: [],
  keepDirStruct: true,
  openCache: true,
  mkdirIfSingle: false,
  deleteDir: false,
}`

const defaultConfig = {
  set(data: string) {
    localStorage.setItem(DEFAULT_CONFIG, data)
  },
  get() {
    return localStorage.getItem(DEFAULT_CONFIG) || DEFAULT_TEMPLATE
  },
}

export default defaultConfig
