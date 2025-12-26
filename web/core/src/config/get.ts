import path from 'path'
import { pathExists } from 'fs-extra'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const get = async (configPath: string): Promise<any> => {
    const resolvedPath = path.isAbsolute(configPath)
        ? configPath
        : path.resolve(process.cwd(), configPath)
    if (!(await pathExists(resolvedPath))) {
        return
    }
    try {
        const config = await import(resolvedPath)
        return config.default
    } catch (e) {
        console.error('配置文件加载失败', e)
        return
    }
}

export default get
