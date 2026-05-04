import { Button } from "@workspace/ui/components/button"
import { useEffect, useRef, useState } from "react"

import {
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from "@workspace/ui/components/tabs"

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@workspace/ui/components/tooltip"
import { ScrollArea } from "@workspace/ui/components/scroll-area"
import { Separator } from "@workspace/ui/components/separator"
import { Textarea } from "@workspace/ui/components/textarea"

import {
  Mic,
  MicOff,
  Video,
  VideoOff,
  Monitor,
  Users,
  PhoneOff,
  MessageSquare,
  Send,
} from "lucide-react"

const DUMMY_MESSAGES = [
  { id: 1, type: "system", text: "Meeting started · 23:20" },
  {
    id: 2,
    type: "message",
    from: "Alex Johnson",
    colorClass: "text-blue-400",
    time: "23:21",
    text: "Hey everyone, can you all hear me okay?",
    self: false,
  },
  {
    id: 3,
    type: "message",
    from: "Sarah R.",
    colorClass: "text-amber-400",
    time: "23:21",
    text: "Yep, loud and clear! 👍",
    self: false,
  },
  {
    id: 4,
    type: "message",
    from: "Mike K.",
    colorClass: "text-zinc-400",
    time: "23:22",
    text: "Same here. Starting screen share now.",
    self: false,
  },
  { id: 5, type: "system", text: "Mike K. started screen sharing" },
  {
    id: 6,
    type: "message",
    from: "Alex Johnson",
    colorClass: "text-blue-400",
    time: "23:25",
    text: "Client 1 has different track IDs between signaling and OnTrack — checking SDP renegotiation",
    self: false,
  },
  {
    id: 7,
    type: "message",
    from: "You",
    colorClass: "text-emerald-400",
    time: "23:26",
    text: "Client 1 triggered a second offer before the first was acked. I'll debounce negotiationneeded.",
    self: true,
  },
  {
    id: 8,
    type: "message",
    from: "Sarah R.",
    colorClass: "text-amber-400",
    time: "23:26",
    text: "Makes sense. Stream group 12 is the camera track, not screen share.",
    self: false,
  },
]

const PARTICIPANTS = [
  {
    id: 1,
    initials: "AJ",
    name: "Alex Johnson",
    role: "Host",
    color: "bg-blue-600",
    muted: false,
  },
  {
    id: 2,
    initials: "SR",
    name: "Sarah R.",
    role: null,
    color: "bg-teal-700",
    muted: true,
  },
  {
    id: 3,
    initials: "MK",
    name: "Mike K.",
    role: null,
    color: "bg-orange-700",
    muted: false,
  },
  {
    id: 4,
    initials: "Me",
    name: "You",
    role: "me",
    color: "bg-violet-700",
    muted: false,
  },
]

function ParticipantRow({
  participant,
}: {
  participant: (typeof PARTICIPANTS)[number]
}) {
  return (
    <div className="flex items-center gap-3 rounded-md px-3 py-2 transition-colors hover:bg-zinc-700/40">
      <div
        className={`h-8 w-8 rounded-full ${participant.color} flex shrink-0 items-center justify-center text-xs font-medium text-white`}
      >
        {participant.initials}
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm text-zinc-200">{participant.name}</p>
        {participant.role === "Host" && (
          <p className="text-[11px] text-blue-400">Host</p>
        )}
      </div>
      {participant.muted && (
        <MicOff className="h-3.5 w-3.5 shrink-0 text-red-400" />
      )}
      {!participant.muted && (
        <Mic className="h-3.5 w-3.5 shrink-0 text-zinc-500" />
      )}
    </div>
  )
}

function ChatMessage({ msg }: { msg: (typeof DUMMY_MESSAGES)[number] }) {
  if (msg.type === "system") {
    return (
      <div className="py-0.5 text-center text-[11px] text-zinc-500">
        {msg.text}
      </div>
    )
  }

  return (
    <div
      className={`flex flex-col gap-0.5 ${msg.self ? "items-end" : "items-start"}`}
    >
      <div
        className={`flex items-baseline gap-1.5 ${msg.self ? "flex-row-reverse" : ""}`}
      >
        <span className={`text-[11px] font-medium ${msg.colorClass}`}>
          {msg.from}
        </span>
        <span className="text-[10px] text-zinc-600">{msg.time}</span>
      </div>
      <div
        className={`max-w-[90%] px-2.5 py-1.5 text-[12.5px] leading-relaxed break-words ${
          msg.self
            ? "rounded-lg rounded-tr-none bg-blue-900/60 text-zinc-100"
            : "rounded-lg rounded-tl-none bg-zinc-700/60 text-zinc-200"
        }`}
      >
        {msg.text}
      </div>
    </div>
  )
}

const ChatTabs = () => {
  const bottomRef = useRef<HTMLDivElement | null>(null)
  const [messages, setMessages] = useState(DUMMY_MESSAGES)
  const [input, setInput] = useState("")

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  function sendMessage() {
    const text = input.trim()
    if (!text) return
    const now = new Date()
    const time = `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`
    setMessages((prev) => [
      ...prev,
      {
        id: Date.now(),
        type: "message",
        from: "You",
        colorClass: "text-emerald-400",
        time,
        text,
        self: true,
      },
    ])
    setInput("")
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  return (
    <div className="flex h-screen w-64 flex-col border-l border-zinc-700/50 bg-[#1a1a28]/95 xl:w-xl">
      <Tabs defaultValue="chat" className="flex min-h-0 flex-1 flex-col">
        {/* Tab headers */}
        <div className="border-b border-zinc-700/50 px-3 pt-2">
          <TabsList className="h-8 w-full bg-zinc-700/40">
            <TabsTrigger
              value="chat"
              className="flex-1 text-xs data-[state=active]:bg-blue-600 data-[state=active]:text-white data-[state=inactive]:text-zinc-400"
            >
              <MessageSquare className="mr-1 h-3 w-3" />
              Chat
            </TabsTrigger>
            <TabsTrigger
              value="participants"
              className="flex-1 text-xs data-[state=active]:bg-blue-600 data-[state=active]:text-white data-[state=inactive]:text-zinc-400"
            >
              <Users className="mr-1 h-3 w-3" />
              People ({PARTICIPANTS.length})
            </TabsTrigger>
          </TabsList>
        </div>

        {/* Chat tab */}
        <TabsContent
          value="chat"
          className="mt-0 flex h-full min-h-0 flex-1 flex-col"
        >
          <div className="flex-1 overflow-y-auto px-3 py-2">
            <div className="flex flex-col gap-3">
              {messages.map((msg) => (
                <ChatMessage key={msg.id} msg={msg} />
              ))}
              <div ref={bottomRef} />
            </div>
          </div>

          <Separator className="bg-zinc-700/50" />

          <div className="flex items-end gap-2 p-2">
            <Textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Message everyone..."
              className="max-h-20 min-h-[36px] flex-1 resize-none rounded-2xl border-zinc-600/50 bg-zinc-700/50 px-3 py-2 text-xs text-zinc-100 placeholder:text-zinc-500 focus-visible:ring-blue-500"
              rows={1}
            />
            <Button
              size="icon"
              onClick={sendMessage}
              disabled={!input.trim()}
              className="h-8 w-8 shrink-0 rounded-full bg-blue-600 hover:bg-blue-500 disabled:opacity-30"
            >
              <Send className="h-3.5 w-3.5" />
            </Button>
          </div>
        </TabsContent>

        {/* Participants tab */}
        <TabsContent value="participants" className="mt-0 min-h-0 flex-1">
          <ScrollArea className="h-full px-2 py-2">
            <div className="flex flex-col gap-0.5">
              {PARTICIPANTS.map((p) => (
                <ParticipantRow key={p.id} participant={p} />
              ))}
            </div>
          </ScrollArea>
        </TabsContent>
      </Tabs>
    </div>
  )
}

export default ChatTabs
