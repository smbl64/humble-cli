package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/smbl64/humble-cli/internal/commands"
	"github.com/smbl64/humble-cli/internal/config"
	"github.com/smbl64/humble-cli/internal/models"
	"github.com/smbl64/humble-cli/internal/util"
	"github.com/spf13/cobra"
)

var version = "dev" // Set by build

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "humble-cli: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "humble-cli",
	Short: "The missing Humble Bundle CLI",
	Long:  "Command-line tool to interact with Humble Bundle purchases: list bundles, show details, search products, and download items.",
}

var authCmd = &cobra.Command{
	Use:   "auth <SESSION-KEY>",
	Short: "Set the authentication session key",
	Long: `Set the session key used for authentication with Humble Bundle API.
See online documentation on how to find the session key from your web browser.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.SetConfig(args[0])
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all your purchased bundles",
	RunE: func(cmd *cobra.Command, args []string) error {
		fields, _ := cmd.Flags().GetStringSlice("field")
		claimed, _ := cmd.Flags().GetString("claimed")

		// Normalize fields to lowercase
		for i := range fields {
			fields[i] = strings.ToLower(fields[i])
		}

		return commands.ListBundles(fields, claimed)
	},
}

var listChoicesCmd = &cobra.Command{
	Use:   "list-choices",
	Short: "List your current Humble Choices",
	RunE: func(cmd *cobra.Command, args []string) error {
		periodStr, _ := cmd.Flags().GetString("period")
		period := models.NewChoicePeriod(periodStr)
		return commands.ListHumbleChoices(period)
	},
}

var detailsCmd = &cobra.Command{
	Use:     "details <BUNDLE-KEY>",
	Aliases: []string{"info"},
	Short:   "Print details of a certain bundle",
	Long:    "Print details of a certain bundle. The key can be partially entered.",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.ShowBundleDetails(args[0])
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <KEYWORDS...>",
	Short: "Search through all bundle products for keywords",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keywords := strings.Join(args, " ")
		modeStr, _ := cmd.Flags().GetString("mode")

		mode, err := models.ParseMatchMode(modeStr)
		if err != nil {
			return err
		}

		return commands.Search(keywords, mode)
	},
}

var downloadCmd = &cobra.Command{
	Use:     "download <BUNDLE-KEY>",
	Aliases: []string{"d"},
	Short:   "Selectively download items from a bundle",
	Long:    "Selectively download items from a bundle. The key can be partially entered.",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		formats, _ := cmd.Flags().GetStringSlice("format")
		maxSizeStr, _ := cmd.Flags().GetString("max-size")
		itemNumbers, _ := cmd.Flags().GetString("item-numbers")
		torrentsOnly, _ := cmd.Flags().GetBool("torrents")
		curDir, _ := cmd.Flags().GetBool("cur-dir")

		// Normalize formats to lowercase
		for i := range formats {
			formats[i] = strings.ToLower(formats[i])
		}

		var maxSize uint64
		if maxSizeStr != "" {
			size, err := util.ByteStringToNumber(maxSizeStr)
			if err != nil {
				return fmt.Errorf("failed to parse the specified size: %s: %w", maxSizeStr, err)
			}
			maxSize = size
		}

		return commands.DownloadBundle(args[0], formats, maxSize, itemNumbers, torrentsOnly, curDir)
	},
}

var bulkDownloadCmd = &cobra.Command{
	Use:     "bulk-download <INPUT-FILE>",
	Aliases: []string{"b"},
	Short:   "Download items from multiple bundles",
	Long: `Download items from multiple bundles listed in an input file.
This takes the input created from the list command, then iterates all items,
using the bundle name as directory name.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		formats, _ := cmd.Flags().GetStringSlice("format")
		maxSizeStr, _ := cmd.Flags().GetString("max-size")
		torrentsOnly, _ := cmd.Flags().GetBool("torrents")
		curDir, _ := cmd.Flags().GetBool("cur-dir")

		// Normalize formats to lowercase
		for i := range formats {
			formats[i] = strings.ToLower(formats[i])
		}

		var maxSize uint64
		if maxSizeStr != "" {
			size, err := util.ByteStringToNumber(maxSizeStr)
			if err != nil {
				return fmt.Errorf("failed to parse the specified size: %s: %w", maxSizeStr, err)
			}
			maxSize = size
		}

		return commands.DownloadBundles(args[0], formats, maxSize, torrentsOnly, curDir)
	},
}

func init() {
	rootCmd.Version = version

	// list command flags
	listCmd.Flags().StringSliceP("field", "", []string{}, "Print bundle with the specified fields only (key, name, size, claimed)")
	listCmd.Flags().StringP("claimed", "", "all", "Show claimed or unclaimed bundles only (all, yes, no)")

	// list-choices command flags
	listChoicesCmd.Flags().StringP("period", "", "current", "The month and year to use (e.g., 'january-2023' or 'current')")

	// search command flags
	searchCmd.Flags().StringP("mode", "", "any", "Whether all or any of the keywords should match (all, any)")

	// download command flags
	downloadCmd.Flags().StringP("item-numbers", "i", "", "Download only specified items (comma-separated, supports ranges like '1-5' or '10-')")
	downloadCmd.Flags().StringSliceP("format", "f", []string{}, "Filter downloaded items by format (can be specified multiple times)")
	downloadCmd.Flags().BoolP("torrents", "t", false, "Download only .torrent files for items")
	downloadCmd.Flags().StringP("max-size", "s", "", "Filter downloaded items by maximum size (e.g., 14MB or 4GiB)")
	downloadCmd.Flags().BoolP("cur-dir", "c", false, "Download into current directory (no bundle directory created)")

	// bulk-download command flags
	bulkDownloadCmd.Flags().StringSliceP("format", "f", []string{}, "Filter downloaded items by format (can be specified multiple times)")
	bulkDownloadCmd.Flags().BoolP("torrents", "t", false, "Download only .torrent files for items")
	bulkDownloadCmd.Flags().StringP("max-size", "s", "", "Filter downloaded items by maximum size (e.g., 14MB or 4GiB)")
	bulkDownloadCmd.Flags().BoolP("cur-dir", "c", false, "Download into current directory (no bundle directory created)")

	// Add all subcommands
	rootCmd.AddCommand(
		authCmd,
		listCmd,
		listChoicesCmd,
		detailsCmd,
		searchCmd,
		downloadCmd,
		bulkDownloadCmd,
	)
}
