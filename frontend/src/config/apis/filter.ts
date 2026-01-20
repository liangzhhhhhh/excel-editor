import {types} from "../../../wailsjs/go/models";
import CommonResponse = types.CommonResponse;
import {errorLogout} from "@/views/utils";
import {Message} from "@arco-design/web-vue";

export enum RespCode {
    NormalCode      = 20000,
    NoInitedCode    = 30000,
    AuthCode        = 40000,
    ErrorCode       = 50000,
    LocalFileNoFoundCode = 50001,
    FileReadCode    = 50002,
    ConsistentCheckCode = 50003,
    ConsistentCheckNoInitedCode = 50004,
}

export function handleResp<T>(resp: CommonResponse): T {
    switch (resp.status) {
        case RespCode.NormalCode:
            return resp.data
        case RespCode.NoInitedCode:
            throw new Error('内网数据未拉取成功', { cause: resp.status })
        case RespCode.AuthCode:
            errorLogout()
            throw new Error('未登录')
        case RespCode.ErrorCode:
        case RespCode.FileReadCode:
        case RespCode.LocalFileNoFoundCode:
            throw new Error(resp.msg)
        case RespCode.ConsistentCheckCode:
        case RespCode.ConsistentCheckNoInitedCode:
            throw new Error(resp.msg, { cause: resp.status })
        default:
            throw new Error(resp.msg || '请求失败')
    }
}

