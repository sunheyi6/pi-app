/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// Wails Go 后端方法类型声明
declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          StartAgent(cwd: string, sessionPath: string): Promise<void>
          StopAgent(): void
          SendPrompt(message: string, images: any[]): Promise<string>
          SendSteer(message: string): Promise<string>
          SendFollowUp(message: string): Promise<string>
          Abort(): Promise<string>
          SetModel(provider: string, modelId: string): Promise<string>
          SetThinkingLevel(level: string): Promise<string>
          GetState(): Promise<string>
          GetMessages(): Promise<string>
          NewSession(): Promise<string>
          GetSessions(): Promise<string>
          GetForks(): Promise<string>
          Fork(entryId: string): Promise<string>
          ExportHTML(outputPath: string): Promise<string>
          GetAvailableModels(): Promise<string>
          GetCommands(): Promise<string>
          SelectDirectory(): Promise<string>
          GetAppInfo(): Promise<string>
        }
      }
    }
    runtime?: {
      EventsOn(eventName: string, callback: (...args: any[]) => void): void
      EventsOff(eventName: string): void
    }
  }
}

export {}
