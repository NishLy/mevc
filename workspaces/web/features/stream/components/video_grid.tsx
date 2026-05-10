"use client"

import classNames from "classnames"
import { useDeferredValue } from "react"
import VideoTile from "./video_tile"
import useMeet from "../state/meet"
import { Button } from "@/components/ui/button"
import { ArrowLeftFromLine } from "lucide-react"
import { AnimatePresence, motion } from "framer-motion"
import PersistentPiP from "@/components/pip"

const calculateGridColumns = (count: number) => {
  if (count === 1 || count === 0)
    return "grid-cols-1 md:grid-cols-1 lg:grid-cols-1 xl:max-w-9/12"
  if (count === 2)
    return "grid-cols-2 md:grid-cols-2 lg:grid-cols-2 xl:max-w-10/12"
  if (count <= 4)
    return "grid-cols-2 md:grid-cols-2 lg:grid-cols-2 xl:max-w-10/12"
  if (count <= 6)
    return "grid-cols-2 md:grid-cols-3 lg:grid-cols-3 xl:max-w-12/12"
  if (count <= 9)
    return "grid-cols-2 md:grid-cols-3 lg:grid-cols-3 xl:max-w-10/12"
  if (count <= 16)
    return "grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-4 xl:max-w-12/12"
  if (count <= 25) return "grid-cols-3 md:grid-cols-4 lg:grid-cols-5"
  return "grid-cols-3 md:grid-cols-4 lg:grid-cols-6"
}

export default function VideosGrid() {
  const rtcService = useMeet((state) => state.RTCService)
  const localStreams = useMeet((state) => state.localStreams)
  const { currentPage, totalPages, visibleStreams, maxiumPerPage } = useMeet(
    (state) => state.pagination
  )

  const pinnedStreams = useDeferredValue(
    visibleStreams.filter((s) =>
      useMeet.getState().pinnedStreamIds.includes(s?.id ?? "")
    )
  )

  const unpinnedStreams = useDeferredValue(
    visibleStreams.filter(
      (s) => s && !useMeet.getState().pinnedStreamIds.includes(s?.id ?? "")
    )
  )

  const unpinnedClas = useDeferredValue(
    calculateGridColumns(
      totalPages > 1 ? maxiumPerPage : unpinnedStreams.length
    )
  )

  return (
    <div className="relative flex h-full w-full flex-col items-center justify-center">
      <div
        className={classNames(
          "box-border w-full content-center gap-2 bg-transparent p-2",
          pinnedStreams.length === 0 && unpinnedClas,
          pinnedStreams.length > 0
            ? "flex w-screen justify-items-start gap-2 overflow-x-auto overflow-y-hidden p-4"
            : "mx-auto grid w-full justify-center justify-items-center",
          unpinnedStreams.length > 1 && pinnedStreams.length > 0
            ? "h-[15vh]"
            : ""
        )}
      >
        <AnimatePresence mode="popLayout">
          {localStreams
            .filter((s) => !!s)
            .map((s) =>
              unpinnedStreams.length > 0 ? (
                <PersistentPiP key={s.id} className="z-40 aspect-video">
                  <div className="h-full w-full">
                    <VideoTile props={s} />
                  </div>
                </PersistentPiP>
              ) : (
                <div className="h-full w-full" key={s.id}>
                  <VideoTile props={s} />
                </div>
              )
            )}

          {visibleStreams.map((s, index) => (
            <motion.div
              layout // <--- This prevents the immediate snap/flicker of neighbors
              key={index}
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.8, transition: { duration: 0.2 } }}
              className={classNames(
                pinnedStreams.length > 0
                  ? "relative z-30 h-40 w-70 shrink-0 opacity-70 hover:opacity-100"
                  : "relative h-fit w-full",
                !s && "hidden"
              )}
            >
              <VideoTile props={s} />
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      <div className="absolute bottom-26 flex w-fit items-center justify-center gap-4 rounded-2xl bg-white/20 px-4 py-2 text-sm text-white">
        <Button
          variant="outline"
          size="icon"
          disabled={currentPage === 1}
          onClick={() => {
            if (currentPage > 1) {
              const prevPage = currentPage - 1
              rtcService?.requestPageChange(prevPage)
              useMeet.setState((state) => ({
                pagination: {
                  ...state.pagination,
                  currentPage: prevPage,
                },
              }))
            }
          }}
          className="cursor-pointer transition-all"
        >
          <ArrowLeftFromLine size="20" className="" />
        </Button>
        <span className="min-w-20 text-center font-semibold">
          {currentPage} / {totalPages}
        </span>
        <Button
          variant="outline"
          size="icon"
          disabled={currentPage === totalPages}
          onClick={() => {
            if (currentPage < totalPages) {
              const nextPage = currentPage + 1
              rtcService?.requestPageChange(nextPage)
              useMeet.setState((state) => ({
                pagination: {
                  ...state.pagination,
                  currentPage: nextPage,
                },
              }))
            }
          }}
          className="cursor-pointer transition-all"
        >
          <ArrowLeftFromLine size="20" className="rotate-180" />
        </Button>
      </div>
    </div>
  )
}
