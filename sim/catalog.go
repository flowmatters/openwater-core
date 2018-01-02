package sim

var Catalog = make(map[string] (func() TimeSteppingModel))


// Need to expand to be a catalog entry for the model, including metadata, etc
