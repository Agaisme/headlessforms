/**
 * API Client - Centralized API abstraction layer
 * Implements Repository pattern for all backend API calls
 */

import { auth } from '$lib/stores/auth';
import { get } from 'svelte/store';

// Types for API responses
export interface ApiResponse<T> {
	status: 'success' | 'error';
	data?: T;
	message?: string;
	error?: string;
}

export interface PaginatedResponse<T> {
	items: T[];
	pagination: {
		page: number;
		limit: number;
		total: number;
		total_pages: number;
	};
}

// Base API client
class ApiClient {
	private baseUrl = '/api/v1';

	private async request<T>(
		endpoint: string,
		options: RequestInit = {}
	): Promise<ApiResponse<T>> {
		const authStore = get(auth);
		const token = authStore.token;

		const headers: HeadersInit = {
			'Content-Type': 'application/json',
			...options.headers
		};

		if (token) {
			(headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
		}

		try {
			const response = await fetch(`${this.baseUrl}${endpoint}`, {
				...options,
				headers
			});

			const json = await response.json();

			if (!response.ok) {
				return {
					status: 'error',
					message: json.message || json.error || 'Request failed',
					error: json.error
				};
			}

			return {
				status: 'success',
				data: json.data
			};
		} catch (error) {
			return {
				status: 'error',
				message: error instanceof Error ? error.message : 'Network error'
			};
		}
	}

	get<T>(endpoint: string): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, { method: 'GET' });
	}

	post<T>(endpoint: string, data?: unknown): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'POST',
			body: data ? JSON.stringify(data) : undefined
		});
	}

	put<T>(endpoint: string, data?: unknown): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'PUT',
			body: data ? JSON.stringify(data) : undefined
		});
	}

	delete<T>(endpoint: string): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, { method: 'DELETE' });
	}
}

export const api = new ApiClient();
