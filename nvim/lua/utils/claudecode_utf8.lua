local M = {}

local notified = {}
local patched_functions = {}

local function is_continuation(byte)
  return byte and byte >= 0x80 and byte <= 0xBF
end

local function expected_sequence_length(first_byte)
  if first_byte >= 0xC2 and first_byte <= 0xDF then
    return 2
  elseif first_byte >= 0xE0 and first_byte <= 0xEF then
    return 3
  elseif first_byte >= 0xF0 and first_byte <= 0xF4 then
    return 4
  end

  return nil
end

local function is_valid_second_byte(first_byte, second_byte)
  if not second_byte then
    return false
  elseif first_byte == 0xE0 then
    return second_byte >= 0xA0 and second_byte <= 0xBF
  elseif first_byte == 0xED then
    return second_byte >= 0x80 and second_byte <= 0x9F
  elseif first_byte == 0xF0 then
    return second_byte >= 0x90 and second_byte <= 0xBF
  elseif first_byte == 0xF4 then
    return second_byte >= 0x80 and second_byte <= 0x8F
  end

  return is_continuation(second_byte)
end

local function valid_sequence_length(text, index)
  local first_byte = text:byte(index)
  if not first_byte then
    return nil
  elseif first_byte <= 0x7F then
    return 1
  end

  local length = expected_sequence_length(first_byte)
  if not length or not is_valid_second_byte(first_byte, text:byte(index + 1)) then
    return nil
  end

  for offset = 2, length - 1 do
    if not is_continuation(text:byte(index + offset)) then
      return nil
    end
  end

  return length
end

---@param text string
---@return boolean
function M.is_valid(text)
  if type(text) ~= "string" then
    return false
  end

  local index = 1
  while index <= #text do
    local length = valid_sequence_length(text, index)
    if not length then
      return false
    end
    index = index + length
  end

  return true
end

---@param text string
---@return string
function M.sanitize(text)
  if type(text) ~= "string" then
    return ""
  end

  local valid_parts = {}
  local index = 1
  while index <= #text do
    local length = valid_sequence_length(text, index)
    if length then
      valid_parts[#valid_parts + 1] = text:sub(index, index + length - 1)
      index = index + length
    else
      index = index + 1
    end
  end

  return table.concat(valid_parts)
end

local function truncated_suffix_start(text)
  local index = 1
  while index <= #text do
    local length = valid_sequence_length(text, index)
    if length then
      index = index + length
    else
      local first_byte = text:byte(index)
      local expected_length = first_byte and expected_sequence_length(first_byte)
      local remaining = #text - index + 1

      if not expected_length or remaining >= expected_length then
        return nil
      end

      if remaining >= 2 and not is_valid_second_byte(first_byte, text:byte(index + 1)) then
        return nil
      end

      for offset = 2, remaining - 1 do
        if not is_continuation(text:byte(index + offset)) then
          return nil
        end
      end

      return index
    end
  end

  return nil
end

local function replace_text(selection, text)
  selection.text = text
  if text == "" and type(selection.selection) == "table" then
    selection.selection.isEmpty = true
  end
end

---@param selection table|nil
---@return table|nil
function M.repair_selection(selection)
  if type(selection) ~= "table" or type(selection.text) ~= "string" or M.is_valid(selection.text) then
    return selection
  end

  local suffix_start = truncated_suffix_start(selection.text)
  if not suffix_start then
    return selection
  end

  local valid_prefix = selection.text:sub(1, suffix_start - 1)
  local selected_suffix = selection.text:sub(suffix_start)
  local range = selection.selection
  local range_end = type(range) == "table" and range["end"] or nil
  local line_number = type(range_end) == "table" and range_end.line or nil
  local byte_index = type(range_end) == "table" and range_end.character or nil

  if
    type(line_number) == "number"
    and line_number >= 0
    and line_number == math.floor(line_number)
    and type(byte_index) == "number"
    and byte_index == math.floor(byte_index)
  then
    local lines_ok, lines = pcall(vim.api.nvim_buf_get_lines, 0, line_number, line_number + 1, false)
    local line = lines_ok and lines[1] or nil

    if type(line) == "string" and byte_index >= 1 and byte_index <= #line then
      local start_ok, start_offset = pcall(vim.str_utf_start, line, byte_index)
      local end_ok, end_offset = pcall(vim.str_utf_end, line, byte_index)

      if start_ok and end_ok and type(start_offset) == "number" and type(end_offset) == "number" then
        local character_start = byte_index + start_offset
        local character_end = byte_index + end_offset

        if
          character_start >= 1
          and character_start <= byte_index
          and character_end >= byte_index
          and character_end <= #line
        then
          local selected_line_bytes = line:sub(character_start, byte_index)
          local full_character = line:sub(character_start, character_end)
          local repaired = valid_prefix .. full_character

          if selected_suffix == selected_line_bytes and M.is_valid(full_character) and M.is_valid(repaired) then
            replace_text(selection, repaired)
            return selection
          end
        end
      end
    end
  end

  replace_text(selection, valid_prefix)
  return selection
end

---@param key string
---@param message string
function M.notify_once(key, message)
  if notified[key] then
    return
  end

  notified[key] = true
  vim.notify(message, vim.log.levels.WARN)
end

---@param original function
---@return function
function M.wrap_selection_result(original)
  return function(...)
    return M.repair_selection(original(...))
  end
end

---@param original function
---@return function
function M.wrap_send_selection_update(original)
  return function(selection, ...)
    if type(selection) == "table" and type(selection.text) == "string" and not M.is_valid(selection.text) then
      replace_text(selection, M.sanitize(selection.text))
      M.notify_once(
        "sanitized_selection",
        "claudecode.nvim: invalid UTF-8 was removed from a selection before broadcast"
      )
    end

    return original(selection, ...)
  end
end

---@param function_name string
---@param wrapper fun(original: function): function
---@return boolean patched
function M.patch_selection_function(function_name, wrapper)
  local selection_ok, selection = pcall(require, "claudecode.selection")
  local original = selection_ok and type(selection) == "table" and selection[function_name] or nil

  if type(original) ~= "function" then
    M.notify_once(
      "missing_function:" .. function_name,
      string.format("claudecode.nvim UTF-8 patch skipped: selection.%s is unavailable", function_name)
    )
    return false
  end

  if patched_functions[function_name] == original then
    return true
  end

  local wrapped = wrapper(original)
  selection[function_name] = wrapped
  patched_functions[function_name] = wrapped
  return true
end

return M
