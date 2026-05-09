"use client"

import classNames from "classnames"
import { useDeferredValue, useMemo } from "react"
import VideoTile from "./video_tile"
import useMeet from "../state/meet"
import { Button } from "@/components/ui/button"
import { ArrowLeftFromLine } from "lucide-react"
import { AnimatePresence, motion } from "framer-motion"
import PersistentPiP from "@/components/pip"

const calculateGridColumns = (count: number) => {
  if (count === 1)
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
  const localStreams = useMeet((state) => state.localStreams)
  const remoteStreams = useMeet((state) => state.remoteStreams)
  const roomState = useMeet((state) => state.roomState)
  const currentPage = useMeet((state) => state.currentPage)
  const totalPages = Math.ceil(
    roomState.current_total_grouped_streams / roomState.maxium_per_page
  )

  const streams = useMemo(
    () => [...localStreams, ...remoteStreams].filter((s) => !!s),
    [localStreams, remoteStreams]
  )

  const pinnedStreams = useDeferredValue(
    streams.filter((s) => useMeet.getState().pinnedStreamIds.includes(s.id))
  )
  const unpinnedStreams = useDeferredValue(
    streams.filter((s) => !useMeet.getState().pinnedStreamIds.includes(s.id))
  )

  // const skeletonArray = Array.from({ length: roomState.maxium_per_page }, (_, i) => i)

  // prevent layout shift when pinning/unpinning by keeping the grid structure consistent
  const pinnedClas = useDeferredValue(
    calculateGridColumns(pinnedStreams.length)
  )
  const unpinnedClas = useDeferredValue(
    calculateGridColumns(unpinnedStreams.length)
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
          {/* {localStreams
            .filter((s) => !!s)
            .map((s) => (
              <PersistentPiP key={s.id} className="z-40 aspect-video">
                <div className="h-full w-full">
                  <VideoTile {...s} />
                </div>
              </PersistentPiP>
            ))} */}

          {unpinnedStreams.map((s) => (
            <motion.div
              layout // <--- This prevents the immediate snap/flicker of neighbors
              key={s.id}
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.8, transition: { duration: 0.2 } }}
              className={classNames(
                pinnedStreams.length > 0
                  ? "relative z-30 h-40 w-70 shrink-0 opacity-70 hover:opacity-100"
                  : "relative h-fit w-full"
              )}
            >
              <VideoTile {...s} />
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      {/* Pinned Section */}
      <AnimatePresence mode="popLayout">
        {pinnedStreams.length > 0 && (
          <motion.div
            layout
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 20 }}
            className={classNames(
              "mx-auto grid h-[90vh] w-full items-center justify-items-center gap-4 overflow-hidden rounded-lg",
              pinnedClas
            )}
          >
            {pinnedStreams.map((s) => (
              <motion.div
                layout // <--- Makes pinned items slide smoothly when others are added/removed
                key={s.id}
                className="relative h-fit w-full"
              >
                <VideoTile {...s} />
              </motion.div>
            ))}
          </motion.div>
        )}
      </AnimatePresence>

      <div className="absolute bottom-26 flex w-fit items-center justify-center gap-4 rounded-2xl bg-white/20 px-4 py-2 text-sm text-white">
        <Button
          variant="outline"
          size="icon"
          disabled={currentPage === 1}
          onClick={() =>
            useMeet.setState((state) => ({
              currentPage: state.currentPage - 1,
            }))
          }
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
          onClick={() =>
            useMeet.setState((state) => ({
              currentPage: state.currentPage + 1,
            }))
          }
          className="cursor-pointer transition-all"
        >
          <ArrowLeftFromLine size="20" className="rotate-180" />
        </Button>
      </div>
    </div>
  )
}
