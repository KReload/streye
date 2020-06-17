<script>
  import { onMount } from "svelte";
  import { tick } from 'svelte';
  import { Confirm } from 'svelte-confirm'
  import { scrollto } from "svelte-scrollto";

  let streams = [];
  onMount(async () => {
    await fetch(`/api/streams`)
      .then(r => r.json())
      .then(data => {
        streams = data;
      });
  });

  let data = {
    name: "",
  };

  let addStream = () => {
    const newStream = {
      name: data.name,
    };
    fetch("/api/stream", {
      method: "POST",
      body: JSON.stringify(newStream),
      headers: { "Content-type": "application/json; charset=UTF-8" }
    })
      .then(response => response.json())
      .then(stream => {
        streams = streams.concat(stream);
        data = {
          id: null,
          name: "",
          playUrl: "",
          paths: [],
        };
      });
  };

  let deleteStream = id => {
    fetch("/api/streams/" + id, { method: "DELETE" }).then(() => {
      streams = streams.filter(stream => stream.id !== id);
    });
  };

  let deleteVideo = (stream, path) => {
    fetch("/api/streams/"+stream.id+"/"+path, { method: "DELETE" }).then(() => {
      stream.paths = stream.paths.filter(currentPath => currentPath !== path);

      // Rerendering
      streams = streams;
    });
  }

  let isEdit = false;
  let cancelEdit = () => {
    isEdit = false;
  }

  let editStream = stream => {
    isEdit = true;
    data = stream;
  };
  
  let updateStream = () => {
    isEdit = false;
    let streamDB = data
    fetch("/api/stream", {
      method: "PATCH",
      body: JSON.stringify(streamDB),
      headers: { "Content-type": "application/json; charset=UTF-8" }
    }).then(() => {
      let objIndex = streams.findIndex(stream => stream.id == streamDB.id);
      streams[objIndex] = streamDB;
      data = {
        id: null,
        name: "",
        playUrl: "",
        paths: [],
      };
    });
  };

  let areaDom = [];
  async function copyToClipboard(text, i) {
    let valueCopy = text;
    areaDom[i].focus();
    areaDom[i].select();
    await tick();
    document.execCommand('copy');
  }
</script>

<style>
  * {
    font-family: inherit;
    font-size: inherit;
  }

  input {
    display: block;
    margin: 0 0 0.5em 0;
  }

  select {
    float: left;
    margin: 0 1em 1em 0;
    width: 14em;
  }

  .buttons {
    clear: both;
  }
</style>

<section>
  <div class="container">
    <div class="row mt-5">
      <div class="col-md-12">
        <p class="h1"><img src="/favicon.png" alt="streye logo">STREYE</p>
      </div>
    </div>
    <div class="row mt-5">
      <div class="col-md-6">
        <div class="card p-2 shadow mb-3">
          <div class="card-body" id="add-stream">
            {#if isEdit === false}
              <h5 class="card-title mb-4" >Add new stream</h5>
            {:else}
              <h5 class="card-title mb-4" >Edit stream</h5>
            {/if}
            <form>
              <div class="form-group">
                <label for="title">Stream name</label>
                <input
                  bind:value={data.name}
                  type="text"
                  class="form-control"
                  id="text"
                  placeholder="Name" />
              </div>
              {#if isEdit === false}
                <button
                  type="submit"
                  on:click|preventDefault={addStream}
                  class="btn btn-primary">
                  Add stream
                </button>
              {:else}
                <button
                  type="submit"
                  on:click|preventDefault={updateStream}
                  class="btn btn-info">
                  Edit Note
                </button>
                <button
                  type="submit"
                  on:click|preventDefault={cancelEdit}
                  class="btn btn-secondary">
                  Cancel
                </button>
              {/if}
            </form>
          </div>
        </div>
        <div class="card p-2 shadow">
          <div class="card-body">
            <p class="h2">Instructions</p>
            <ul>
              <li><div class="h4">Step 1:</div> Create stream</li>
              <li><div class="h4">Step 2:</div> Copy the token</li>
              <li><div class="h4">Step 3:</div> Go to streaming software and fill url with: rtmp://ip-address/live and key stream with given token</li>
              <li><div class="h4">Step 4:</div> Reload page when stream is done to get download link</li>
            </ul>
          </div>
        </div>
      </div>
      <div class="col-md-6">
        {#each streams as stream, i}
          <div class="card mb-3">
            <div class="card-header">{stream.name}</div>
            <div class="card-body">
              {#if stream.token != null}
              <p class="h3">Token</p>
              <div class="input-group">
                <input type="text" class="form-control" readonly="readonly" bind:this={areaDom[i]} bind:value={stream.token}>
                <div class="input-group-append">
                  <button class="btn btn-primary" type="button" on:click={copyToClipboard(stream.token, i)}><i class="fa fa-link" aria-hidden="true"></i> Copy</button>
                </div>
              </div>
              {/if}

              {#if stream.playUrl}
              <hr/>
              <p class="h3">Play uri</p>
              <div>/{stream.playUrl}.m3u8</div>
              {/if}
              
              {#if stream.paths}
                <hr/>
                <div class="card mb-5">
                  <ul class="list-group list-group-flush">
                {#each stream.paths as path}
                
                    <li class="list-group-item">
                      <a href="/rec/{path}">{path}</a>
                      <Confirm
                        let:confirm="{confirmThis}"
                        themeColor="0"
                      >
                        <button class="btn btn-default" type="button" on:click={() => confirmThis(deleteVideo, stream, path)}>
                          <i class="fa fa-trash" aria-hidden="true"></i>
                        </button>

                        <span slot="title">
                          Delete video
                        </span>
                        <span slot="description">
                          Are you sure you want to delete this video ?
                        </span>
                      </Confirm>
                    </li>
                    
                {/each}
                  </ul>
                </div>
              {/if}
              <hr/>
              <button use:scrollto={'#add-stream'} class="btn btn-info" on:click={() => {editStream(stream)}}>
                Edit
              </button>
              <Confirm
                        let:confirm="{confirmThis}"
                        themeColor="0"
              >
                <button class="btn btn-danger" on:click={() => confirmThis(deleteStream, stream.id)}>
                  Delete
                </button>

                <span slot="title">
                  Delete stream
                </span>
                <span slot="description">
                  Are you sure you want to delete this stream ? You should download all saved videos before deleting.
                </span>
              </Confirm>
            </div>
          </div>
        {/each}
      </div>
    </div>
  </div>
</section>
