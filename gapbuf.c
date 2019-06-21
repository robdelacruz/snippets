#include <stdlib.h>
#include <stdio.h>
#include <string.h>

typedef struct gapbuf_t {
    char *bytes;            // total buffer storage including pre, gap, and post segments.
    int bytes_len;          // number of bytes in buffer storage.
    int gap_start, gap_end; // starting and ending indexes to gap.

    // bytes:       [0]...[bytes_len-1] 
    // gap:         [gap_start]...[gap_end]
    // pre:         [0]...[gap_start-1]
    // post:        [gap_end+1]...[bytes_len-1]
    //
    // gap_len =    gap_end-gap_start+1
    // pre_len =    gap_start
    // post_len =   bytes_len-gap_end-1 

} gapbuf_t;

#define GAP_LEN(g)      (g->gap_end - g->gap_start + 1)
#define PRE_LEN(g)      (g->gap_start)
#define POST_LEN(g)     (g->bytes_len - g->gap_end - 1)
#define PRE_START(g)    (0)
#define PRE_END(g)      (g->gap_start-1)
#define POST_START(g)   (g->gap_end+1)
#define POST_END(g)     (g->bytes_len-1)

gapbuf_t* gapbuf_new() {
    int buf_initial_size = 50;

    gapbuf_t* g = malloc(sizeof(gapbuf_t));
    g->bytes = malloc(buf_initial_size);
    g->bytes_len = buf_initial_size;
    g->gap_start = 0;
    g->gap_end = buf_initial_size-1;

    return g;
}

void gapbuf_free(gapbuf_t* g) {
    free(g->bytes);
    free(g);
}

char* gapbuf_text(gapbuf_t* g) {
    int pre_len = PRE_LEN(g);
    int post_len = POST_LEN(g);

    // Combine the pre and post segments into a single string.
    char *text = malloc(pre_len + post_len + 1);
    if (pre_len > 0) {
        memcpy(text, g->bytes, pre_len);
    }
    if (post_len > 0) {
        memcpy(text+pre_len, g->bytes+POST_START(g), post_len);
    }
    text[pre_len + post_len] = '\0';

    return text;
}

void gapbuf_repr(gapbuf_t* g) {
    printf("bytes len: %d, gap: %d - %d, len: %d\n", g->bytes_len, g->gap_start, g->gap_end, g->gap_end-g->gap_start+1);
    char *text = gapbuf_text(g);
    printf("text: '%s'\n", text);
    free(text);
}

// Return target exponential buffer growth cap.
int buf_cap(int len) {
    int cap = 1;
    while (cap < len) {
        cap *= 2;
    }
    return cap;
}

void _gapbuf_realloc_bytes(gapbuf_t *g, int new_bytes_len) {
    if (new_bytes_len <= g->bytes_len) return;

    // Increase capacity of buffer.
    g->bytes = realloc(g->bytes, new_bytes_len);

    // Shift post segment to the end of buffer.
    int new_gap_end = new_bytes_len - POST_LEN(g) -1;
    memcpy(g->bytes + new_gap_end + 1, g->bytes + g->gap_end + 1, POST_LEN(g));

    g->bytes_len = new_bytes_len;
    g->gap_end = new_gap_end;
}

void gapbuf_insert_text(gapbuf_t* g, char *text) {
    int text_len = strlen(text);
    if (text_len > GAP_LEN(g)) {
        int new_bytes_len = buf_cap(g->bytes_len + text_len);
        _gapbuf_realloc_bytes(g, new_bytes_len);
    }

    // Insert text into gap
    // Then shift gap to the right to be ready for next insert.
    memcpy(g->bytes + g->gap_start, text, text_len);
    g->gap_start += text_len;
}


int main() {
    gapbuf_t* g = gapbuf_new();
    gapbuf_repr(g);

    for (int i=0; i < 20; i++) {
        gapbuf_insert_text(g, "abcdef ");
        gapbuf_repr(g);
    }

    gapbuf_free(g);

/*
    g.bytes = "0123456789";
    g.bytes = "abcde___12345";
    g.bytes = "abcde___12345";
    g.bytes_len = strlen(g.bytes);
    g.gap_start = 5;
    g.gap_end = 7;
    printf("gapbuf text: '%s'\n", gapbuf_text(g));
*/

    return 0;
}

