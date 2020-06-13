<script>
  import { onMount } from "svelte";

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
    twitchToken: ""
  };

  let addStream = () => {
    const newStream = {
      name: data.name,
      twitchToken: data.twitchToken
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
          twitchToken: ""
        };
      });
  };

  let deleteStream = id => {
    fetch("/api/streams/" + id, { method: "DELETE" }).then(() => {
      streams = streams.filter(stream => stream.id !== id);
    });
  };

  let isEdit = false;
  let editStream = stream => {
    isEdit = true;
    data = stream;
  };

  let updateStream = () => {
    isEdit = !isEdit;
    let streamDB = {
      name: data.name,
      twitchToken: data.twitchToken,
      id: data.id,
      token: data.token
    };
    fetch("/api/stream", {
      method: "PATCH",
      body: JSON.stringify(streamDB),
      headers: { "Content-type": "application/json; charset=UTF-8" }
    }).then(() => {
      let objIndex = streams.findIndex(stream => stream.id == streamDB.id);
      streams[objIndex] = streamDB;
      data = {
        id: null,
        title: "",
        category: "",
        content: ""
      };
    });
  };
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

<!-- https://eugenkiss.github.io/7guis/tasks#crud -->
<section>
  <div class="container">
    <div class="row mt-5">
      <div class="col-md-12">
        <h1>STREYE</h1>
      </div>
    </div>
    <div class="row mt-5">
      <div class="col-md-6">
        <div class="card p-2 shadow">
          <div class="card-body">
            <h5 class="card-title mb-4">Add new stream</h5>
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
              <div class="form-group">
                <label for="category">Twitch key (optional)</label>
                <input
                  bind:value={data.twitchToken}
                  type="text"
                  class="form-control"
                  id="text"
                  placeholder="Twitch key" />
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
              {/if}
            </form>
          </div>
        </div>
      </div>
      <div class="col-md-6">
        {#each streams as stream}
          <div class="card mb-3">
            <div class="card-header">{stream.id}</div>
            <div class="card-body">
              <h5 class="card-title">{stream.name}</h5>
              <p class="card-text">{stream.token}</p>
              <p class="card-text">{stream.twitchToken}</p>
              <button class="btn btn-info" on:click={editStream(stream)}>
                Edit
              </button>
              <button class="btn btn-danger" on:click={deleteStream(stream.id)}>
                Delete
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  </div>
</section>
