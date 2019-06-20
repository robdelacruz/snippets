#include <stdlib.h>
#include <string.h>
#include <stdio.h>

typedef struct str_t {
    char *bytes;
    int len;
    int cap;
} str_t;

typedef struct strnode_t {
    str_t* str;
    struct strnode_t *next;
} strnode_t;

// Return target exponential buffer growth cap.
int buf_cap(int len) {
    int cap = 1;
    while (cap < len) {
        cap *= 2;
    }
    return cap;
}

str_t* str_new(char *sz) {
    str_t* str = malloc(sizeof(str_t));
    int len = strlen(sz);
    int cap = buf_cap(len+1);
    str->bytes = malloc(cap);

    strncpy(str->bytes, sz, len);
    str->bytes[len] = '\0';
    str->len = len;
    str->cap = cap;
}

void str_free(str_t* str) {
    free(str->bytes);
    free(str);
}

// Print internal representation of str.
void str_repr(str_t* str) {
    printf("'%s' len:%d cap:%d\n", str->bytes, str->len, str->cap);
}

void str_realloc_len(str_t *str, int new_len) {
    // Check if we need to expand capacity.
    // Set aside 1 byte for '\0' terminator.
    if (new_len+1 > str->cap) {
        int new_cap = buf_cap(new_len+1);
        str->bytes = realloc(str->bytes, new_cap);
        str->cap = new_cap;
    }
    str->len = new_len;
}

void str_append(str_t* str, char *sz) {
    int sz_len = strlen(sz);
    str_realloc_len(str, str->len + sz_len);

    strncat(str->bytes, sz, sz_len);
}

void str_insert(str_t *str, int pos, char *sz) {
    if (pos < 0) {
        pos = 0;
    }
    if (pos > str->len) {
        pos = str->len;
    }

    int sz_len = strlen(sz);
    str_realloc_len(str, str->len + sz_len);

    // Shift original text to the right starting from pos.
    memcpy(str->bytes + pos + sz_len, str->bytes + pos, str->len - pos);

    // Insert new text in pos.
    memcpy(str->bytes + pos, sz, sz_len);
    str->bytes[str->len] = '\0';
}

// Return slice [start:end] comprised of characters from index start to end-1.
str_t* str_slice(str_t *str, int start, int end) {
    if (start < 0) {
        start = 0;
    }
    if (end > str->len) {
        end = str->len;
    }
    if (end <= start) {
        end = start;
    }

    int sz_len = end - start;
    char* slice_sz = malloc(sz_len+1);
    memcpy(slice_sz, str->bytes + start, sz_len);
    slice_sz[sz_len] = '\0';

    str_t *slice_str = str_new(slice_sz);
    free(slice_sz);
    return slice_str;
}

strnode_t* strnode_new(str_t* str) {
    strnode_t* node = malloc(sizeof(strnode_t));
    node->str = str;
    node->next = NULL;
    return node;
}

void strnode_free(strnode_t *head) {
    strnode_t* node = head;
    while (node != NULL) {
        str_free(node->str);
        node->str = NULL;

        strnode_t *tmp = node;
        node = node->next;
        free(tmp);
    }
}

void strnode_repr(strnode_t* head) {
    printf("[\n");

    strnode_t* node = head;
    while (node != NULL) {
        printf("\t");
        str_repr(node->str);

        node = node->next;
    }

    printf("]\n");
}

// Split a str by delim char and return linked list of tokens.
strnode_t* str_split(str_t *str, char delim) {
    strnode_t* head = NULL;
    strnode_t* last_node = NULL;

    int token_start = 0;
    for (int i=0; i < str->len; i++) {
        if (str->bytes[i] == delim) {
            str_t* tokstr = str_slice(str, token_start, i);
            strnode_t* node = strnode_new(tokstr);

            token_start = i+1;

            if (last_node == NULL) {
                head = node;
                last_node = node;
                continue;
            }

            // Append token.
            last_node->next = node;
            last_node = node;
        }
    }

    // Add remaining token.
    str_t* tokstr = str_slice(str, token_start, str->len+1);
    strnode_t* node = strnode_new(tokstr);

    if (last_node == NULL) {
        head = node;
    } else {
        last_node->next = node;
    }

    return head;
}


int main() {
    str_t* s1 = str_new("");
    str_t* s2 = str_new("rob");

    str_repr(s1);
    str_repr(s2);

    for (int i=0; i < 10; i++) {
        str_append(s2, "123_");
        str_repr(s2);
    }

    str_free(s1);
    str_free(s2);

    s1 = str_new("HelloWorld");
    str_insert(s1, 5, "123");
    str_repr(s1);

    s2 = str_new("abcdefghijklmnopqrstuvwxyz");
    str_t* s3 = str_slice(s2, 24, 27);
    str_t* s4 = str_slice(s2, 2, 23);
    str_repr(s2);
    str_repr(s3);
    str_repr(s4);

    str_free(s1);
    str_free(s2);
    str_free(s3);

    s1 = str_new("abc;def;ghi;jklmnop;qrstuv;wxyz");
    strnode_t* toks = str_split(s1, ';');
    strnode_repr(toks);

    strnode_free(toks);
    str_free(s1);

    return 0;
}

