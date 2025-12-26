import fs from 'fs-extra'
import { cachePath } from './paths.js'

export const cacheRecord = {
    read: (): string[] => {
        try {
            if (fs.existsSync(cachePath)) {
                return fs.readJSONSync(cachePath)
            }
        } catch (e) {
            // ignore
        }
        return []
    },
    write: (data: string[]) => {
        try {
            fs.writeJSONSync(cachePath, data)
        } catch (e) {
            // ignore
        }
    }
}

export const saveCache = (files: string[]) => {
    const current = cacheRecord.read()
    const newSet = new Set([...current, ...files])
    cacheRecord.write(Array.from(newSet))
}
