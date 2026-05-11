"use client"

import { useSuspenseQuery } from "@tanstack/react-query"
import { getRoomByCode } from "../services/api"
import PreviewRoom from "./preview/preview"
import { useState } from "react"
import Room from "./room"

interface IPreloadProps {
  code: string
}

export default function Preload({ code }: IPreloadProps) {
  const { data } = useSuspenseQuery({
    queryKey: ["room", code],
    queryFn: () => getRoomByCode(code),
  })

  const [hasDonePreload, setHasDonePreload] = useState(false)

  const handleJoin = () => {
    setHasDonePreload(true)
  }

  return !hasDonePreload ? (
    <PreviewRoom data={data.data.data} onJoin={handleJoin} />
  ) : (
    <Room data={data.data.data} />
  )
}
