interface IGenericResponse<T> {
  success: boolean
  data: T
  message?: string
}

export default IGenericResponse
