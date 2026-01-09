/**
 * Forms Repository - API abstraction for form operations
 */

import { api, type ApiResponse } from './client';

export interface Form {
	id: string;
	public_id: string;
	name: string;
	description?: string;
	owner_id: string;
	access_mode: 'public' | 'with_key' | 'private';
	submission_key?: string;
	redirect_url?: string;
	webhook_url?: string;
	notify_emails?: string[];
	enabled: boolean;
	submission_count?: number;
	unread_count?: number;
	created_at: string;
	updated_at: string;
}

export interface CreateFormInput {
	name: string;
	description?: string;
	access_mode?: 'public' | 'with_key' | 'private';
	redirect_url?: string;
	webhook_url?: string;
	notify_emails?: string[];
}

export interface UpdateFormInput extends Partial<CreateFormInput> {
	enabled?: boolean;
}

export const formsRepository = {
	/**
	 * List all forms with pagination
	 */
	list: async (page = 1, limit = 100) => {
		return api.get<{ forms: Form[]; pagination: unknown }>(`/forms?page=${page}&limit=${limit}`);
	},

	/**
	 * Get a single form by ID
	 */
	get: async (formId: string) => {
		return api.get<Form>(`/forms/${formId}`);
	},

	/**
	 * Create a new form
	 */
	create: async (data: CreateFormInput) => {
		return api.post<Form>('/forms', data);
	},

	/**
	 * Update an existing form
	 */
	update: async (formId: string, data: UpdateFormInput) => {
		return api.put<Form>(`/forms/${formId}`, data);
	},

	/**
	 * Delete a form
	 */
	delete: async (formId: string) => {
		return api.delete<void>(`/forms/${formId}`);
	},

	/**
	 * Get form statistics
	 */
	getStats: async (formId: string) => {
		return api.get<unknown>(`/forms/${formId}/stats`);
	}
};
