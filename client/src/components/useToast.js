import { useCallback, useState } from "react";

export function useToast() {
  const [toast, setToast] = useState(null);

  const show = useCallback((title, message, ms = 2500) => {
    setToast({ title, message });
    window.clearTimeout(show._t);
    show._t = window.setTimeout(() => setToast(null), ms);
  }, []);

  const clear = useCallback(() => setToast(null), []);

  return { toast, show, clear };
}
