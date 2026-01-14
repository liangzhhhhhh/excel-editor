import { ElLoading, ElMessage } from 'element-plus'
import {handleResp} from "@/config/apis/filter";
import {types} from "../../../wailsjs/go/models";
import CommonResponse = types.CommonResponse;

let loading: any
let loadingCount = 0

export interface CallOptions {
    loading?: boolean
    silent?: boolean
}

// @ts-ignore
export async function runApi<T>(
    api: () => Promise<CommonResponse<T>>,
    options: CallOptions = {}
): Promise<T> {
    const {
        loading: useLoading = true,
        silent = false
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
        throw err

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


