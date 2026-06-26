import { useSearchParams } from "react-router-dom";

export type ViewState = "default" | "loading" | "empty" | "error" | "denied" | "success" | "long" | "focus";

export function useViewState(): ViewState {
  const [params] = useSearchParams();
  const state = params.get("state");
  if (state === "loading" || state === "empty" || state === "error" || state === "denied" || state === "success" || state === "long" || state === "focus") {
    return state;
  }
  return "default";
}
