import queryString from 'query-string'

type TAllowMethods = 'get' | 'post' | 'put' | 'delete'

type ResponseType<T> = {
  success: boolean
  errorMessage?: string
  data?: T
}

type TFetch = <T>(url: string, params?: Record<string, any>) => Promise<T>

const TOKEN_KEY = 'linker_auth_token'

const allowMethods: TAllowMethods[] = ['get', 'delete', 'post', 'put']

const fetch = allowMethods.reduce((result, method) => {
  result[method] = async <T>(url: string, params?: Record<string, any>) => {
    // Get auth token from localStorage
    const token = localStorage.getItem(TOKEN_KEY)

    const headers: Record<string, string> = {
      'content-type': 'application/json',
    }

    // Add Authorization header if token exists
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const fetchOption: RequestInit = {
      method,
      headers,
    }
    if ((method === 'get' || method === 'delete') && params) {
      // Use arrayFormat 'none' to serialize arrays as key=value&key=value2
      // This is compatible with gin's c.QueryArray()
      url += `?${queryString.stringify(params, { arrayFormat: 'none' })}`
    } else if (params) {
      fetchOption.body = JSON.stringify(params)
    }

    const response = await window.fetch(url, fetchOption)
    const responseData = (await response.json()) as ResponseType<T>

    // Handle 401 Unauthorized - clear auth and redirect to login
    if (response.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem('linker_username')
      window.location.href = '/login'
      throw new Error(responseData.errorMessage || '认证失败，请重新登录')
    }

    if (responseData.success) {
      return responseData.data!
    }
    throw new Error(responseData.errorMessage)
  }
  return result
  // @ts-ignore
}, {} as Record<TAllowMethods, TFetch>)

export default fetch

