import api from "@/lib/api"
import IGenericResponse from "@/types/response"
import IRoom from "../types"

export const getRoomByCode = async (code: string) => {
  return await api.get<IGenericResponse<IRoom>>(`rooms/code/${code}`)
}
