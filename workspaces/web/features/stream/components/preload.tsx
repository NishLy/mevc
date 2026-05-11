"use client"

import { useSuspenseQuery } from "@tanstack/react-query"
import { getRoomByCode } from "../services/api"
import PreviewRoom from "./preview/preview"

interface IPreloadProps {
  code: string
}

export default function Preload({ code }: IPreloadProps) {
  const { data } = useSuspenseQuery({
    queryKey: ["room", code],
    queryFn: () => getRoomByCode(code),
  })

  return <PreviewRoom data={data.data.data} />
}
