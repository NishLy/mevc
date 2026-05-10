import { Button } from "@/components/ui/button"
import { use, useEffect, useRef, useState } from "react"

import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"

import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { Textarea } from "@/components/ui/textarea"

import { Mic, MicOff, Users, MessageSquare, Send } from "lucide-react"
import useMeet from "../state/meet"
import { generateInitials, truncateString } from "@/lib/strings"
import ColorGen from "@/lib/color"
import { ChatMessage, ParticipantData } from "../types/service"
import { AnimatePresence, motion } from "framer-motion"
import { useOnClickOutside } from "@/hooks/touch"

function ParticipantRow({
  participant,
}: {
  participant: ParticipantData & { initials: string; color: string }
}) {
  return (
    <div className="flex items-center gap-3 rounded-md px-3 py-2 transition-colors hover:bg-zinc-700/40">
      <div
        className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-medium text-white`}
        style={{ backgroundColor: participant.color }}
      >
        {participant.initials}
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm text-zinc-200">{participant.username}</p>
        {participant.role === "Host" && (
          <p className="text-[11px] text-blue-400">Host</p>
        )}
      </div>
      {participant.isMuted && (
        <MicOff className="h-3.5 w-3.5 shrink-0 text-red-400" />
      )}
      {!participant.isMuted && (
        <Mic className="h-3.5 w-3.5 shrink-0 text-zinc-500" />
      )}
    </div>
  )
}

import React from "react"

export const renderMessageWithLinks = (text: string) => {
  const urlRegex = /(https?:\/\/[^\s]+|www\.[^\s]+)/gi

  const parts = text.split(urlRegex)

  return parts.map((part, index) => {
    // If the part is undefined (can happen with split), skip it
    if (!part) return null

    // Check if the current part is a URL
    if (part.match(urlRegex)) {
      const href = part.startsWith("www.") ? `https://${part}` : part

      return (
        <a
          key={index}
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          className="break-all text-blue-300 hover:underline"
        >
          {part}
        </a>
      )
    }

    // Otherwise, return the plain text
    return <span key={index}>{part}</span>
  })
}

function ChatRow({ msg }: { msg: ChatMessage }) {
  if (msg.type === "system") {
    return (
      <div className="py-0.5 text-center text-[11px] text-zinc-500">
        {msg.message}
      </div>
    )
  }

  return (
    <div
      className={`flex flex-col gap-0.5 ${msg.senderId === "you" ? "items-end" : "items-start"}`}
    >
      <div
        className={`flex items-baseline gap-1.5 ${msg.senderId === "you" ? "flex-row-reverse" : ""}`}
      >
        <span className={`text-[11px] font-medium`}>{msg.senderName}</span>
        <span className="text-[10px] text-zinc-400">
          {new Date(msg.timestamp).toLocaleTimeString()}
        </span>
      </div>
      <div
        className={`max-w-[90%] px-2.5 py-1.5 text-[12.5px] leading-relaxed whitespace-pre-wrap ${
          msg.senderId === "you"
            ? "rounded-lg rounded-tr-none bg-blue-900/60 text-zinc-100"
            : "rounded-lg rounded-tl-none bg-zinc-700/60 text-zinc-200"
        }`}
      >
        {renderMessageWithLinks(msg.message)}
      </div>
    </div>
  )
}

const ChatTabs = () => {
  const scrollContainerRef = useRef<HTMLDivElement | null>(null)
  const bottomRef = useRef<HTMLDivElement | null>(null)
  const messages = useMeet((state) => state.chatMessages)
  const [input, setInput] = useState("")
  const { isChatOpen, isParticipantsOpen } = useMeet(
    (state) => state.uiControls
  )
  const containerRef = useRef<HTMLDivElement | null>(null)
  const participants = useMeet((state) => state.participants)
  const rtcService = useMeet((state) => state.RTCService)
  const isAllFetched = useMeet((state) => state.isChatAllFetched)

  const closeAll = () => {
    useMeet.setState((state) => ({
      uiControls: {
        ...state.uiControls,
        isChatOpen: false,
        isParticipantsOpen: false,
      },
    }))
  }

  useOnClickOutside(containerRef as React.RefObject<HTMLDivElement>, closeAll)
  const isOpen = isChatOpen || isParticipantsOpen

  const memoizedParticipants = participants.map((p) => ({
    ...p,
    initials: generateInitials(p.username),
    color: ColorGen.next(),
  }))

  function sendMessage() {
    const text = truncateString(input.trim(), 1000)
    if (!text) return

    rtcService?.sendChatMessage(text)
    setInput("")

    const message: ChatMessage = {
      senderId: "you",
      senderName: "You",
      timestamp: Date.now(),
      type: "text",
      message: text,
    }

    useMeet.setState((state) => ({
      chatMessages: [...state.chatMessages, message],
    }))

    bottomRef.current?.scrollIntoView({
      behavior: "smooth",
      block: "end",
    })
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  const historySentinelRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!scrollContainerRef.current || isAllFetched) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && rtcService) {
          rtcService.requestChatHistory(messages.length)
        }
      },
      {
        root: scrollContainerRef.current,
        rootMargin: "100px 0px 0px 0px",
        threshold: 0,
        scrollMargin: "100px",
      }
    )

    if (historySentinelRef.current) {
      observer.observe(historySentinelRef.current)
    }

    return () => observer.disconnect()
  }, [isAllFetched, messages.length, rtcService, scrollContainerRef])

  useEffect(() => {
    if (bottomRef.current) {
      bottomRef.current.scrollIntoView({
        behavior: "smooth",
        block: "end",
      })
    }
  }, [bottomRef])

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Optional: Dark backdrop that fades in */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-40 bg-black/20 backdrop-blur-[2px]"
          />

          {/* The Sidebar */}
          <motion.div
            ref={containerRef}
            initial={{ x: "100%" }} // Start off-screen
            animate={{ x: 0 }} // Slide in
            exit={{ x: "100%" }} // Slide out
            transition={{ type: "spring", damping: 25, stiffness: 200 }}
            className="fixed top-0 right-0 bottom-0 z-50 flex h-screen w-64 flex-col border-l border-zinc-700/50 bg-[#1a1a28]/95 xl:w-xl"
          >
            <Tabs
              defaultValue="chat"
              className="flex min-h-0 flex-1 flex-col"
              value={
                isChatOpen
                  ? "chat"
                  : isParticipantsOpen
                    ? "participants"
                    : undefined
              }
            >
              {/* Tab headers */}
              <div className="border-b border-zinc-700/50 p-2 px-3">
                <TabsList className="h-8 w-full bg-zinc-700/40">
                  <TabsTrigger
                    value="chat"
                    className="flex-1 cursor-pointer text-xs data-[state=active]:bg-blue-600 data-[state=active]:text-white data-[state=inactive]:text-zinc-400"
                    onClick={() => {
                      useMeet.setState((state) => ({
                        uiControls: {
                          ...state.uiControls,
                          isChatOpen: !state.uiControls.isChatOpen,
                          isParticipantsOpen: false,
                        },
                      }))
                    }}
                  >
                    <MessageSquare className="mr-1 h-3 w-3" />
                    Chat
                  </TabsTrigger>
                  <TabsTrigger
                    value="participants"
                    className="flex-1 cursor-pointer text-xs data-[state=active]:bg-blue-600 data-[state=active]:text-white data-[state=inactive]:text-zinc-400"
                    onClick={() => {
                      useMeet.setState((state) => ({
                        uiControls: {
                          ...state.uiControls,
                          isParticipantsOpen:
                            !state.uiControls.isParticipantsOpen,
                          isChatOpen: false,
                        },
                      }))
                    }}
                  >
                    <Users className="mr-1 h-3 w-3" />
                    Participants ({memoizedParticipants.length})
                  </TabsTrigger>
                </TabsList>
              </div>

              {/* Chat tab */}
              <TabsContent
                value="chat"
                className="mt- flex h-full min-h-0 flex-1 flex-col"
              >
                <AnimatePresence initial={false}>
                  <div
                    className="flex-1 overflow-y-auto px-3 py-2"
                    ref={scrollContainerRef}
                  >
                    {/* 1. TOP: Observe this to load OLDER history */}
                    <div ref={historySentinelRef} className="h-4 w-full">
                      {isAllFetched ? (
                        <p className="text-center text-[11px] text-zinc-500">
                          No more messages
                        </p>
                      ) : (
                        <p className="text-center text-[11px] text-zinc-400">
                          Loading history...
                        </p>
                      )}
                    </div>

                    {messages.map((msg, i) => (
                      <motion.div
                        key={i + msg.timestamp}
                        initial={{ opacity: 0, y: 10, scale: 0.95 }}
                        animate={{ opacity: 1, y: 0, scale: 1 }}
                        exit={{ opacity: 0 }}
                        transition={{
                          type: "spring",
                          stiffness: 260,
                          damping: 20,
                        }}
                        className="mb-1"
                      >
                        <ChatRow msg={msg} />
                      </motion.div>
                    ))}

                    <div ref={bottomRef}></div>
                  </div>
                </AnimatePresence>

                <div className="flex items-center gap-2 p-2 pb-4">
                  <Textarea
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder="Message everyone..."
                    className="max-h-20 min-h-9 flex-1 resize-none rounded-2xl border-zinc-600/50 bg-zinc-700/50 px-3 py-2 text-xs text-zinc-100 placeholder:text-zinc-500 focus-visible:ring-blue-500"
                    rows={1}
                    maxLength={1000}
                  />
                  <Button
                    size="icon"
                    onClick={sendMessage}
                    disabled={!input.trim()}
                    className="h-full w-12 shrink-0 rounded-2xl focus-visible:ring-blue-500 disabled:opacity-30"
                  >
                    <Send className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </TabsContent>

              {/* Participants tab */}
              <TabsContent value="participants" className="mt-0 min-h-0 flex-1">
                <ScrollArea className="h-full px-2 py-2">
                  <div className="flex flex-col gap-0.5">
                    {memoizedParticipants.map((p) => (
                      <ParticipantRow key={p.clientId} participant={p} />
                    ))}
                  </div>
                </ScrollArea>
              </TabsContent>
            </Tabs>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}

export default ChatTabs
