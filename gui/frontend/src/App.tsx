import { useState, useEffect } from "react";
import "./App.css";
import { Button } from "./components/ui/button";
import { Card, CardContent } from "./components/ui/card";
import { Badge } from "./components/ui/badge";
import { Terminal, Zap, Code2, ChevronRight } from "lucide-react";

const TERMINAL_COMMANDS = [
  {
    command: "myenv start",
    output: "✓ Welcome! Let's get your environment set up!",
  },
  {
    command: "What dose this app do",
    output: "✓ This app is a setup assistant for myenv",
  },
  {
    command: "Let's start the setup",
    output: "✓ Let's start the setup",
  },
];

function App() {
  const [particles, setParticles] = useState<
    Array<{ id: number; x: number; y: number; delay: number }>
  >([]);
  const [terminalLines, setTerminalLines] = useState<
    Array<{ text: string; type: "command" | "output" }>
  >([]);
  const [currentCommand, setCurrentCommand] = useState(0);
  const [typingText, setTypingText] = useState("");
  const [isTyping, setIsTyping] = useState(false);

  useEffect(() => {
    // Generate random particles for background animation
    const newParticles = Array.from({ length: 40 }, (_, i) => ({
      id: i,
      x: Math.random() * 100,
      y: Math.random() * 100,
      delay: Math.random() * 5,
    }));
    setParticles(newParticles);

    // Start terminal animation after a delay
    const timer = setTimeout(() => {
      runTerminalAnimation();
    }, 1500);

    return () => clearTimeout(timer);
  }, []);

  const runTerminalAnimation = async () => {
    // Reset state at the beginning
    setTerminalLines([]);
    setTypingText("");
    setCurrentCommand(0);
    for (let i = 0; i < TERMINAL_COMMANDS.length; i++) {
      setCurrentCommand(i);
      setIsTyping(true);

      // Type command character by character
      const cmd = TERMINAL_COMMANDS[i].command;
      for (let j = 0; j <= cmd.length; j++) {
        setTypingText(cmd.slice(0, j));
        await new Promise((resolve) => setTimeout(resolve, 50));
      }

      // Wait a bit after typing
      await new Promise((resolve) => setTimeout(resolve, 300));

      // Add command to history
      setTerminalLines((prev) => [...prev, { text: cmd, type: "command" }]);
      setTypingText("");
      setIsTyping(false);

      // Wait before showing output
      await new Promise((resolve) => setTimeout(resolve, 200));

      // Add output
      setTerminalLines((prev) => [
        ...prev,
        { text: TERMINAL_COMMANDS[i].output, type: "output" },
      ]);

      // Wait before next command
      await new Promise((resolve) => setTimeout(resolve, 800));
    }

    // Loop the animation
    setTimeout(() => {
      setTerminalLines([]);
      setTypingText("");
      setCurrentCommand(0);
      runTerminalAnimation();
    }, 3000);
  };

  return (
    <div className="relative min-h-screen w-full overflow-hidden bg-linear-to-br from-slate-950 via-blue-950 to-slate-900">
      {/* Animated Grid Background */}
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-size-[4rem_4rem] mask-[radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)] animate-[grid_20s_linear_infinite]" />

      {/* Floating Particles */}
      {particles.map((particle) => (
        <div
          key={particle.id}
          className="absolute w-1 h-1 bg-cyan-400/30 rounded-full animate-float"
          style={{
            left: `${particle.x}%`,
            top: `${particle.y}%`,
            animationDelay: `${particle.delay}s`,
            animationDuration: `${10 + Math.random() * 10}s`,
          }}
        />
      ))}

      {/* Gradient Orbs */}
      <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/20 rounded-full blur-3xl animate-orb-float-1" />
      <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/20 rounded-full blur-3xl animate-orb-float-2" />

      {/* Main Content */}
      <div className="relative z-10 flex flex-col items-center justify-center min-h-screen px-4">
        {/* Main Card */}
        <Card
          className="w-full max-w-2xl bg-slate-900/40 border-slate-700/50 backdrop-blur-xl shadow-2xl shadow-cyan-500/10 animate-fade-in-up"
          style={{ animationDelay: "0.3s" }}
        >
          <CardContent className="p-12">
            {/* Icon Group */}
            <div className="flex justify-center gap-4 mb-8">
              <div className="p-4 bg-linear-to-br from-cyan-500/20 to-cyan-600/20 rounded-2xl backdrop-blur-sm border border-cyan-500/30 animate-bounce-subtle shadow-lg shadow-cyan-500/20">
                <Terminal className="w-10 h-10 text-cyan-400" />
              </div>
              <div
                className="p-4 bg-linear-to-br from-blue-500/20 to-blue-600/20 rounded-2xl backdrop-blur-sm border border-blue-500/30 animate-bounce-subtle shadow-lg shadow-blue-500/20"
                style={{ animationDelay: "0.2s" }}
              >
                <Code2 className="w-10 h-10 text-blue-400" />
              </div>
              <div
                className="p-4 bg-linear-to-br from-purple-500/20 to-purple-600/20 rounded-2xl backdrop-blur-sm border border-purple-500/30 animate-bounce-subtle shadow-lg shadow-purple-500/20"
                style={{ animationDelay: "0.4s" }}
              >
                <Zap className="w-10 h-10 text-purple-400" />
              </div>
            </div>

            {/* Welcome Text */}
            <h1 className="text-5xl md:text-6xl font-bold text-center mb-4 bg-linear-to-r from-cyan-400 via-blue-400 to-purple-400 bg-clip-text text-transparent">
              Welcome to myenv
            </h1>

            <p className="text-lg text-center text-slate-400 mb-8 leading-relaxed">
              Let's get started with your setup
            </p>

            {/* Embedded Terminal Animation */}
            <div className="mb-8 font-mono text-sm bg-slate-950/60 border border-slate-700/50 rounded-lg overflow-hidden">
              {/* Terminal Header */}
              <div className="flex items-center gap-2 px-4 py-2.5 bg-slate-900/60 border-b border-slate-700/50">
                <Terminal className="w-3.5 h-3.5 text-cyan-400" />
                <span className="text-slate-400 text-xs">myenv</span>
                <div className="flex gap-1.5 ml-auto">
                  <div className="w-2.5 h-2.5 rounded-full bg-red-500/60" />
                  <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/60" />
                  <div className="w-2.5 h-2.5 rounded-full bg-green-500/60" />
                </div>
              </div>

              {/* Terminal Content */}
              <div className="p-4 space-y-1 h-40 overflow-hidden flex flex-col text-left">
                <div className="flex-1 min-h-0">
                  {terminalLines.map((line, index) => (
                    <div
                      key={index}
                      className={
                        line.type === "command"
                          ? "text-cyan-300"
                          : "text-green-400 text-xs"
                      }
                    >
                      {line.type === "command" ? "$ " : "  "}
                      {line.text}
                    </div>
                  ))}
                  {isTyping && (
                    <div className="text-cyan-300">
                      $ {typingText}
                      <span className="inline-block w-2 h-4 bg-cyan-400 ml-0.5 animate-pulse" />
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* CTA Button */}
            <Button
              size="lg"
              className="w-full text-lg py-6 bg-linear-to-r from-cyan-500 to-blue-500 hover:from-cyan-600 hover:to-blue-600 border-0 shadow-lg shadow-cyan-500/30 hover:shadow-cyan-500/50 transition-all duration-300 group"
            >
              Begin Setup
              <ChevronRight className="ml-2 w-5 h-5 transition-transform group-hover:translate-x-1" />
            </Button>

            {/* Footer Note */}
            <p className="text-center text-slate-500 text-xs mt-6">
              This will guide you through the initial configuration
            </p>
          </CardContent>
        </Card>

        {/* Version Badge */}
        <div
          className="absolute bottom-8 right-8 animate-fade-in"
          style={{ animationDelay: "1.2s" }}
        >
          <Badge
            variant="secondary"
            className="bg-slate-800/50 border-slate-700/50 text-slate-400 backdrop-blur-sm"
          >
            myenv v1.0.0
          </Badge>
        </div>
      </div>
    </div>
  );
}

export default App;
