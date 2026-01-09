import { writable } from "svelte/store";
import { browser } from "$app/environment";

interface User {
  id: string;
  email: string;
  name: string;
  role: "super_admin" | "admin" | "user";
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

const initialState: AuthState = {
  user: null,
  token: null,
  isAuthenticated: false,
  isLoading: true,
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  // Load from localStorage on init
  if (browser) {
    const savedToken = localStorage.getItem("auth_token");
    const savedUser = localStorage.getItem("auth_user");

    if (savedToken && savedUser) {
      try {
        const user = JSON.parse(savedUser);
        set({
          user,
          token: savedToken,
          isAuthenticated: true,
          isLoading: false,
        });
      } catch {
        localStorage.removeItem("auth_token");
        localStorage.removeItem("auth_user");
        set({ ...initialState, isLoading: false });
      }
    } else {
      set({ ...initialState, isLoading: false });
    }
  }

  return {
    subscribe,

    login: async (email: string, password: string) => {
      const res = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      const json = await res.json();

      if (json.status === "success") {
        const { token, user } = json.data;

        if (browser) {
          localStorage.setItem("auth_token", token);
          localStorage.setItem("auth_user", JSON.stringify(user));
        }

        set({
          user,
          token,
          isAuthenticated: true,
          isLoading: false,
        });

        return { success: true };
      }

      return { success: false, message: json.message || "Login failed" };
    },

    register: async (email: string, password: string, name: string) => {
      const res = await fetch("/api/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password, name }),
      });

      const json = await res.json();

      if (json.status === "success") {
        const { token, user } = json.data;

        if (browser) {
          localStorage.setItem("auth_token", token);
          localStorage.setItem("auth_user", JSON.stringify(user));
        }

        set({
          user,
          token,
          isAuthenticated: true,
          isLoading: false,
        });

        return { success: true };
      }

      return { success: false, message: json.message || "Registration failed" };
    },

    logout: () => {
      if (browser) {
        localStorage.removeItem("auth_token");
        localStorage.removeItem("auth_user");
      }
      set({ ...initialState, isLoading: false });
    },

    getToken: () => {
      if (browser) {
        return localStorage.getItem("auth_token");
      }
      return null;
    },
  };
}

export const auth = createAuthStore();
