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
}

export function handleResp<T>(resp: CommonResponse): T {
    switch (resp.status) {
        case RespCode.NormalCode:
            return resp.data
        case RespCode.NoInitedCode:
            throw new Error('内网数据未拉取成功')
        case RespCode.AuthCode:
            errorLogout()
            throw new Error('未登录')
        case RespCode.ErrorCode:
        case RespCode.FileReadCode:
        case RespCode.LocalFileNoFoundCode:
            throw new Error(resp.msg)
        default:
            throw new Error(resp.msg || '请求失败')
    }
}

