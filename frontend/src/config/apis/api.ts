import { ElLoading, ElMessage } from 'element-plus'
import {handleResp} from "@/config/apis/filter";
import {types} from "../../../wailsjs/go/models";
import CommonResponse = types.CommonResponse;
import { LogError } from '../../../wailsjs/go/main/App';

let loading: any
let loadingCount = 0

export interface CallOptions {
    loading?: boolean
    silent?: boolean
    throwOnError?: boolean
}

// @ts-ignore
export async function runApi<T>(
    api: () => Promise<CommonResponse>,
    options: CallOptions = {}
): Promise<T> {
    const {
        loading: useLoading = true,
        silent = false,
        throwOnError = true
    } = options

    try {
        if (useLoading) {
            if (loadingCount === 0) {
                loading = ElLoading.service({ lock: true })
            }
            loadingCount++
        }

        const resp = await api()
        return handleResp(resp)

    } catch (err: any) {
        if (!silent) {
            ElMessage.error(err?.message || '请求异常')
        }

        // 写日志到本地
        LogError(err.message || '未知错误', err.stack || '')

        if (throwOnError) {
            throw err
        }

        // 不向上抛时，返回一个“空值”
        return undefined as unknown as T
        // 或者 return undefined as unknown as T

    } finally {
        if (useLoading) {
            loadingCount--
            if (loadingCount <= 0) {
                loading?.close()
                loadingCount = 0
            }
        }
    }
}
