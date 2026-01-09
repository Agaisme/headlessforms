/**
 * API Module - Re-exports all repositories
 */

export { api, type ApiResponse, type PaginatedResponse } from './client';
export { formsRepository, type Form, type CreateFormInput, type UpdateFormInput } from './forms';
export { submissionsRepository, type Submission } from './submissions';
