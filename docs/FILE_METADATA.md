# Version stamps and variable names in HDF5 files

Model files and results files lay their data out by position: the second
output of a model is whatever that model's second output was in the build that
wrote the file. Model descriptions change between builds, so a file read
against the wrong build returns the wrong variable without complaint. Two
pieces of metadata prevent that.

This is a contract between two codebases. openwater-core writes it from Go
(`io/hdf5_attrs.go`, `util/version.go`, `cmd/ow-sim/stamp.go`); the `openwater`
Python package writes and reads it (`openwater/stamps.py`). Change one side and
the other has to follow.

## Version stamp

String attributes at the root of the file.

| attribute | meaning |
| --- | --- |
| `openwater_version` | full version of the openwater-core build the file is laid out for, eg `1.0.3+fe85ae1.6d3bb266` |
| `signature_hash` | that build's model signature hash. Same hash, same layout |
| `created_by` | the program that wrote the file: `openwater-py <version>`, `ow-sim`, `ows-ensemble` |
| `created_timestamp` | UTC, ISO 8601, no zone suffix |
| `model_openwater_version` | results files only: `openwater_version` of the model file the run used |
| `model_signature_hash` | results files only: `signature_hash` of that model file |

Who stamps what:

- **Model files** are stamped by `OWTemplate.write_model()` with the loaded
  library's version. Files derived from a model file (`clip`, `substitute`,
  `split`, `retime`) keep the stamp of the file they came from, because they
  keep its layout.
- **Results files** are stamped by `ow-sim`: the main output file, a separate
  `-final-states` file, and each `-outputs` split file (by the `-writer`
  process that owns it). `ows-ensemble` stamps the file it writes to. The two
  `model_*` attributes are left out when the model file has no stamp.

`ow-sim` logs a warning when the model file's `signature_hash` differs from its
own, and runs anyway.

Values are variable-length UTF-8 strings, which is what h5py writes for
`f.attrs[name] = "text"`, so they read back as `str` in Python whichever side
wrote them.

## Variable names

A `VARIABLES` attribute on each of a model's datasets names the entries along
its variable axis:

| dataset | shape | names |
| --- | --- | --- |
| `/MODELS/<model>/inputs` | (cell, **input**, timestep) | input names |
| `/MODELS/<model>/outputs` | (cell, **output**, timestep) | output names |
| `/MODELS/<model>/states` | (cell, **state**) | state names |
| `/MODELS/<model>/parameters` | (**parameter**, cell) | parameter names; a table parameter has a row per element, named `name[0]`, `name[1]`, ... |

The names come from the model's description (its OW-SPEC) in the build that
wrote the dataset. They are stored as a 1-D array of fixed-length byte strings,
the same layout as the `DIMS` attribute on a model's `map`.

The attribute is only written when the number of names equals the size of the
variable axis. If they differ there is no telling which name belongs where,
and a wrong name is worse than none. Readers apply the same rule and ignore
names that don't fit.

## Reading

`OpenwaterResults` reads a variable by name when the dataset has `VARIABLES`,
whatever build is loaded. A dataset without names is read by position against
the loaded library's description; if the dataset has a different number of
variables than that description, reading raises `VariableMismatchError`
(override with `allow_variable_mismatch=True`).

On open it also compares the results file's `signature_hash` with the loaded
library's and warns once on a difference. `releases.model_file_version()`,
`find_installed_for_model()`, `install_for_model()` and `use_for_model()` accept a results file as well
as a model file, so the matching release can be found and activated from
either.

Files written before any of this existed carry neither stamp nor names, and are
read by position as they always were.
